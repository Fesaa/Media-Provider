# Preferences

Media-Provider allows you to set a few preferences, these can be done via the Preferences tab in settings.
Currently, the following are available.

### Subscription download time
At which hour to start downloading subscriptions, this may be UTC time, depending on your environment.

### Empty subscription download notifications
Subscriptions add a notification when finished, so you can view what has been download during the night
(or day depending on the previous setting). By default, Media-Provider will not add one if nothing was downlaoded
if this has been enabled; a notification will always be added

### Dynasty Tag => Genre
Dynasty has no distinction between Tags and Genres, you may configure some tags (by id or displayed name) to be used as
a genre in the generated ComicInfo.xml. These tags will not be used elsewhere. Does not overwrite the blacklist

### Blacklist tags
Tags configured in this list will not be added as Genre or Tag.
Check the table below for which values you can provide

| Provider  | Mapping                                       |
|-----------|-----------------------------------------------|
| Mangadex  | Checks ID and the used language               |
| Dynasty   | Checks ID and displayed name                  |
| WebToon   | Does not respect the blacklist at this moment |
