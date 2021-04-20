# Information to collect
- Plane ID (Tail number is 2-7 alphanumeric code)
- Timestamp - Unix timestamp (? s/ms precision)
- Current Position
  - Latitude - Valid from -180 to 180, precision of 8 decimals
  - Longitude - Valid from -90 to 90, precision of 8 decimals
  - Altitude - Feet
- Positional Change
  - Airspeed - Knots (? precision)
  - Turn - *Unsure of measurement unit*, direction of bank (side to side)
  - Compass direction - Degrees (? precision)
  - Vertical speed - _Positive for climbing, Negative for descending_ Feet per minute or knots
- Status
  - Attitude (roll and pitch) - degrees (? precision)
  - Heading (should be approx. equal to compass) - Degrees (? precision)
  - Deviation - Degrees and nautical miles (? precision)
  - Action Underway - Enumeration of actions such as taxi, takeoff, in-flight, awaiting landing clearance, landing, etc.

References:
- [How to Read Basic Aircraft Instruments](http://www.actforlibraries.org/how-to-read-basic-aircraft-instruments/)
- [Aircraft Cockpit Instruments Explained for Newbies](http://digitalpilotschool.com/aircraft-cockpit-instruments-explained-for-newbies/)
- [Automatic Dependent Surveillance–Broadcast](https://en.wikipedia.org/wiki/Automatic_Dependent_Surveillance%E2%80%93Broadcast)
- [Decoding ASD-B Packets](https://web.stanford.edu/class/ee179/labs/LabFP_ADSB.html)
- Similar enterprise solution for real flight stream data: [FlightAware](https://flightaware.com/commercial/aeroapi/)


## Valid Longitude and Latitude
From [this post on StackOverflow](https://stackoverflow.com/a/47188298/12676661):
> Valid longitudes are from -180 to 180 degrees.
> Latitudes are supposed to be from -90 degrees to 90 degrees, but areas very near to the poles are not indexable.
> So exact limits, as specified by EPSG:900913 / EPSG:3785 / OSGEO:41001 are the following:
>  -  Valid longitudes are from -180 to 180 degrees.
>  -  Valid latitudes are from -85.05112878 to 85.05112878 degrees.

# Record Size
Each record PUT can be 1kb max (before base64 encoding)
