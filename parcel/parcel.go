package parcel

/*

OSU Google Hackathon 2014
Russell Barnes
Cezary Wojcik
Carly Farr
Sophie Zhu

API:

dropparcel
Update the location of the parcel.
    <server>/dropparcel?longitude=44.12345&latitude=-18.12345
Optionally specify a group (defaults to 1), note and/or userid:
    <server>/dropparcel?longitude=44.12345&latitude=-18.12345&groupid=1&note=your_message_here&userid=userid_string_here

pickupparcel
Archive the parcel's previous location, prevent it from being picked up again.
    <server>/pickupparcel?userid=userid_string_here
Optionally specify a group (defaults to 1):
    <server>/pickupparcel?userid=userid_string_here&groupid=1

locateparcels
Receive a list of all parcels.  Returns a JSON array of parcel objects with the **first** location being the most recent.
    <server>/locateparcels
Optionally specify a group (defaults to 1):
    <server>/locateparcels?groupid=#

seedparcel
Start a new parcel.  This will remove the previous parcel in the group, along with its history.
    <server>/seedparcel?longitude=44.12345&latitude=-18.12345
Optionally specify a group (defaults to 1) and/or note:
    <server>/seedparcel?longitude=44.12345&latitude=-18.12345&groupid=1&note=your_message_here

*/

import (
    "fmt"
    "net/http"
	"encoding/json"
	"appengine"
    "appengine/datastore"
	"time"
)

type Parcel struct {
	Longitude string
	Latitude string
	Groupid string
	Userid string
	Note string
	Active bool
	Date time.Time
}

func init() {
    //http.HandleFunc("/", root)
	http.HandleFunc("/dropparcel", dropparcel)
	http.HandleFunc("/pickupparcel", pickupparcel)
	http.HandleFunc("/locateparcels", locateparcels)
	http.HandleFunc("/seedparcel", seedparcel)
}

func root(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Server is running")
}

func dropparcel(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	
	groupid := r.FormValue("groupid")
	if groupid == "" {
		groupid = "1"
	}
	
	// Ensure that the most recent parcel is not active - find the most recent parcel
	query := datastore.NewQuery("parcelobject").Ancestor(ParentKey(c)).Filter("Groupid =", groupid).Order("-Date").Limit(1)
	for t := query.Run(c); ; {
		var mostrecent Parcel
		_, err := t.Next(&mostrecent)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		
		if mostrecent.Active == true {
			// User needs to pick it up, first!
			fmt.Fprint(w, "[{\"message\":\"failure\"}]")
			return
		}
	}
	
	// Most recent parcel is inactive - add a new parcel to the datastore at the new location
	newparcel := &Parcel{
		Longitude: r.FormValue("longitude"), 
		Latitude: r.FormValue("latitude"),
		Groupid: groupid,
		Userid: r.FormValue("userid"),
		Note: r.FormValue("note"),
		Active: true,
		Date: time.Now(),
	}
	
	
	//c.Infof("newparcel:\n", newparcel)
	
	// format: datastore.NewIncompleteKey(context, "subkind", *parentKey)
	key := datastore.NewIncompleteKey(c, "parcelobject", ParentKey(c))
    if _, err := datastore.Put(c, key, newparcel); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Fprint(w, "[{\"message\":\"failure\"}]")
		return
    }
	fmt.Fprint(w, "[{\"message\":\"success\"}]")
}

func pickupparcel(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	
	// Update the most recent parcel in the datastore
	
	groupid := r.FormValue("groupid")
	if groupid == "" {
		groupid = "1"
	}
	
	// Find the most recent parcel
	query := datastore.NewQuery("parcelobject").Ancestor(ParentKey(c)).Filter("Groupid =", groupid).Order("-Date").Limit(1)
	
	for t := query.Run(c); ; {
		var pickedup Parcel
		key, err := t.Next(&pickedup)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//fmt.Fprint(w, "Key: ", key, "Parcel: ", pickedup)
		
		// Only allow a pick-up if it is still available
		if !pickedup.Active {
			// Parcel has already been picked up
			c.Infof("Pick-up attempted when most recent parcel is already inactive")
			fmt.Fprint(w, "[{\"message\":\"failure\"}]")
			return
		}
		
		pickedup.Active = false
		
		// Update the parcel's availability status in the datastore
		// format: datastore.NewIncompleteKey(context, "subkind", *parentKey)
	    if _, err := datastore.Put(c, key, &pickedup); err != nil {
	        http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	    }
    	fmt.Fprint(w, "[{\"message\":\"success\"}]")   
	}
}

func locateparcels(w http.ResponseWriter, r *http.Request) {
	
	c := appengine.NewContext(r)
	
	
	groupid := r.FormValue("groupid")
	if groupid == "" {
		groupid = "1"
	}
	

	query := datastore.NewQuery("parcelobject").Ancestor(ParentKey(c)).Filter("Groupid =", groupid).Order("-Date")
	query_count, _ := query.Count(c)
	parcels := make([]Parcel, 0, query_count)	// Ten most recent locations returned
	if _, err := query.GetAll(c, &parcels); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	//fmt.Fprint(w, "Parcels:\n", parcels) //users[0].y)
	
	// Respond to the HTML request with JSON-formatted location data
	if parcelbytes, err := json.Marshal(parcels); err != nil {
		c.Infof("Oops - something went wrong with the JSON. \n")
		fmt.Fprint(w, "[{\"message\":\"failure\"}]")
		return
	} else {
		fmt.Fprint(w, string(parcelbytes))	// Print parcel objects in date-descending order as a JSON array
		return
	}
}

func seedparcel(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	
	// Add a parcel to the datastore
	newparcel := &Parcel{
		Longitude: r.FormValue("longitude"), 
		Latitude: r.FormValue("latitude"),
		Groupid: r.FormValue("groupid"),
		Userid: "",
		Note: r.FormValue("note"),
		Active: true,
		Date: time.Now(),
	}
	if newparcel.Groupid == "" {
		newparcel.Groupid = "1"
	}
	
	// If parcels already exists in this group, delete them before adding
	query := datastore.NewQuery("parcelobject").Ancestor(ParentKey(c)).Filter("Groupid =", newparcel.Groupid).Order("-Date")
	
	for t := query.Run(c); ; {
		var x Parcel
		key, err := t.Next(&x)
		if err == datastore.Done {
			c.Infof("Parcels deleted for group: ", newparcel.Groupid)
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//fmt.Fprint(w, "Key: ", key, "Parcel: ", x)
		err = datastore.Delete(c, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	//c.Infof("newparcel:\n", newparcel)
	
	// format: datastore.NewIncompleteKey(context, "subkind", *parentKey)
	key := datastore.NewIncompleteKey(c, "parcelobject", ParentKey(c))
    if _, err := datastore.Put(c, key, newparcel); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }
	fmt.Fprint(w, "[{\"message\":\"success\"}]")
}



// Get the parent key for the particular Parcel entity group
func ParentKey(c appengine.Context) *datastore.Key {
    // The string "development_locationentitygroup" refers to an instance of a LocationEntityGroupType
	// format: datastore.NewKey(context, "groupkind", "groupkind_instance", 0, nil)
    return datastore.NewKey(c, "ParcelEntityGroupType", "development_parcelentitygroup", 0, nil)
}