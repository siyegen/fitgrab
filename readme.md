Grab your data from fitocracy
=============================

## Commands
All commands require ```--username``` and ```--password```. This aren't saved anywhere, but feel free to check the code and confirm!

```bash
./fitgrab --username=$USER --password=$PASS
```
By default running fitgrab will list all exercises (and their count) and also save them in ```$HOMDIR/.fitgrab/YYYY_MM_DD_HH_MM_activities```

### Options

```bash
--stdout
```
This prints the results instead of going to a file

```bash
--file=filename
```
This lets you specify the filename/location

```bash
--name='Backsquat'
```
Grabs all the data for the named workout. This can be quite a lot of data!

```bash
--name-id=174
```
Grabs all the data for the named workout by Fitocracy id. Try this if name isn't working

#### WIP
- Create Fitgrabber object
- create new method for it with optional parameters
- common method for making calls, saving data
- create .store to use for lookups

### TODO
- Add date ranges to queries
- Add option to export / push to keen to get better visuals
- Add option to grab max for a lift / max for number of reps
