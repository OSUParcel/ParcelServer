ParcelServer
============

Google App Engine Server for Oregon State Google Hackathon 2014

Add, update and list parcels using the HTTP GET API below.  All data is returned in JSON format.  Success returns [{"message":"success"}], failure returns [{"message":"failure"}], with the exception of locateparcels returning a JSON array upon success.
##API:

####dropparcel
Update the location of the parcel.  Only succeeds if the parcel was previously inactive.

    <server>/dropparcel?longitude=44.12345&latitude=-18.12345

Optionally specify a group (defaults to 1), userid and/or note:

    <server>/dropparcel?longitude=44.12345&latitude=-18.12345&groupid=1&userid=abcd&note=your_message_here

####pickupparcel
Archive the parcel's previous location, prevent it from being picked up again.

    <server>/pickupparcel

Optionally specify a group (defaults to 1):

    <server>/pickupparcel?groupid=1

####locateparcels
Receive a list of all parcels.  Returns a JSON array of parcel objects with the **first** object being the most recent and the last object the oldest.

    <server>/locateparcels

Optionally specify a group (defaults to 1):

    <server>/locateparcels?groupid=#

####seedparcel
Start a new parcel.  This will remove the previous parcel in the group, along with its history.  Typically only used by administrators to init or reset a parcel group.

    <server>/seedparcel?longitude=44.12345&latitude=-18.12345

Optionally specify a group (defaults to 1) and/or note:

    <server>/seedparcel?longitude=44.12345&latitude=-18.12345&groupid=1&note=your_message_here
