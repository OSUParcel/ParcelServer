ParcelServer
============

Google App Engine Server for Oregon State Google Hackathon 2014

Add, update and list parcels using the HTTP GET API below.
##API:

####dropparcel
Update the location of the parcel.

    <server>/dropparcel?longitude=44.12345&latitude=-106.12345

Optionally specify a group (defaults to 1) and note:

    <server>/dropparcel?longitude=44.12345&latitude=-106.12345&groupid=1&note=your_message_here

####locateparcels
Receive a list of all parcels.  Returns a JSON array of parcel objects with the **first** location being the most recent.

    <server>/locateparcels

Optionally specify a group (defaults to 1):

    <server>/locateparcels?groupid=#

####seedparcel
Start a new parcel.  This will remove the previous parcel in the group, along with its history.

    <server>/seedparcel?longitude=44.12345&latitude=-106.12345

Optionally specify a group (defaults to 1) and note:

    <server>/seedparcel?longitude=44.12345&latitude=-106.12345&groupid=1&note=your_message_here
