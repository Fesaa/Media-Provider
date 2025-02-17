# Dynasty (`dynasty`)

Media-Provider supports downloading manga's from dynasty. Due to the more unregulated nature of the website,
there are currently some limitations, and it is worth checking if the series you're trying to download will work. 

## Tags

Dynasty has a LOT of tags, and most of them are rather weird. By default, dynasty will not embed
any tags into the chapters ComicInfo.xml; you can choose to map some [tags to a genre](./preferences); and
mapping all other tags to the tag field when downloading a series. 

## Config
Dynasty has no additional configuration, and only checks the Query field at the moment.

## Limitations

Dynasty doesn't have a nice API, so series are scrapped from the website; due to the unregulated nature
it's hard to do this correctly for *all* series. The following limitations are currently known

### Might work
- One-shots

### Doesn't work
- Extra chapters without a chapter number
- Some specials might overwrite each other