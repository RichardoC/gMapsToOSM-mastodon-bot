Use the following APIs
 GET /api/v1/notifications
GET /api/v1/notifications/:id
POST /api/v1/notifications/:id/dismiss

Specifically, if there's a "mention" notification, 
take the id, 
find the maps url, 
try to find the coordinates, if this fails
try and make the coorindates by opening the share link and then try to find the corrdinates

If that worked, make the openstreetmap link, then post a reply (somehow?). then dismiss that notification

??? Error cases?


https://docs.joinmastodon.org/methods/notifications/#200-ok