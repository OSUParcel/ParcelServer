package parcel

/*

OSU Google Hackathon 2014
Russell Barnes
Cezary Wojcik
Carly Farr
Sophie Zhu

API:

<server>/seedparcel?longitude=44.12345&latitude=-106.12345&groupid=1

<server>/dropparcel?longitude=44.12345&latitude=-106.12345&groupid=1

<server>/pickupparcel?groupid=1

<server>/locateparcels
<server>/locateparcels?groupid=1

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
	Active bool
	Date time.Time
}

func init() {
    http.HandleFunc("/", root)
	http.HandleFunc("/dropparcel", dropparcel)
	http.HandleFunc("/seedparcel", seedparcel)
	http.HandleFunc("/pickupparcel", pickupparcel)
	http.HandleFunc("/locateparcels", locateparcels)
}

func root(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Server is running")
}

func seedparcel(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	
	// Add a parcel to the datastore
	newparcel := &Parcel{
		Longitude: r.FormValue("longitude"), 
		Latitude: r.FormValue("latitude"),
		Groupid: r.FormValue("groupid"),
		Active: true,
		Date: time.Now(),
	}
	if newparcel.Groupid == "" {
		newparcel.Groupid = "1"
	}
	
	// If parcels already exists in this group, delete them before adding
	query := datastore.NewQuery("parcelobject").Ancestor(ParentKey(c)).Filter("Groupid =", newparcel.Groupid).Order("-Date").Limit(2)
	
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
		//fmt.Fprint(w, "\nParcel deleted\n\n")
		
	}
	
	fmt.Fprint(w, "newparcel:\n", newparcel)
	
	// format: datastore.NewIncompleteKey(context, "subkind", *parentKey)
	key := datastore.NewIncompleteKey(c, "parcelobject", ParentKey(c))
    if _, err := datastore.Put(c, key, newparcel); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
	
}

func dropparcel(w http.ResponseWriter, r *http.Request) {
	// Add a parcel to the datastore
	newparcel := &Parcel{
		Longitude: r.FormValue("longitude"), 
		Latitude: r.FormValue("latitude"),
		Groupid: r.FormValue("groupid"),
		Active: true,
		Date: time.Now(),
	}
	if newparcel.Groupid == "" {
		newparcel.Groupid = "1"
	}
	
	c := appengine.NewContext(r)
	
	fmt.Fprint(w, "newparcel:\n", newparcel)
	
	// format: datastore.NewIncompleteKey(context, "subkind", *parentKey)
	key := datastore.NewIncompleteKey(c, "parcelobject", ParentKey(c))
    if _, err := datastore.Put(c, key, newparcel); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
	
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
		pickedup.Active = false
		if !pickedup.Active {
			fmt.Fprint(w, "the parcel is now inactive\n\n")
		}
		
		// Update the parcel's availability status in the datastore
		// format: datastore.NewIncompleteKey(context, "subkind", *parentKey)
	    if _, err := datastore.Put(c, key, &pickedup); err != nil {
	        http.Error(w, err.Error(), http.StatusInternalServerError)
	    }
	}
}

func locateparcels(w http.ResponseWriter, r *http.Request) {
	
	c := appengine.NewContext(r)
	

	groupid := r.FormValue("groupid")
	if groupid == "" {
		groupid = "1"
	}
	

	query := datastore.NewQuery("parcelobject").Ancestor(ParentKey(c)).Filter("Groupid =", groupid).Order("-Date").Limit(50)
	parcels := make([]Parcel, 0, 10)	// Ten most recent locations returned
	if _, err := query.GetAll(c, &parcels); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	//fmt.Fprint(w, "Parcels:\n", parcels) //users[0].y)
	
	// Respond to the HTML request with JSON-formatted location data
	if parcelbytes, err := json.Marshal(parcels); err != nil {
		fmt.Fprint(w, "Oops - something went wrong with the JSON. \n")
		//fmt.Fprint(w, "{error: 1}")
		return
	} else {
		fmt.Fprint(w, string(parcelbytes))	// Print parcel objects in date-descending order as a JSON array
		return
	}
}



// Get the parent key for the particular Parcel entity group
func ParentKey(c appengine.Context) *datastore.Key {
    // The string "development_locationentitygroup" refers to an instance of a LocationEntityGroupType
	// format: datastore.NewKey(context, "groupkind", "groupkind_instance", 0, nil)
    return datastore.NewKey(c, "ParcelEntityGroupType", "development_parcelentitygroup", 0, nil)
}