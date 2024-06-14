# Mangadex (`mangadex`)

Media-Provider supports downloading manga's from mangadex, currently it will always download the en version, and does no checking if a full volumes can be made. Planning to change this in the future.

You may add as many manga's as you want, but only one will be downloaded at a time, the rest is added in a queue. They will automatically start downloading after. Once a manga's has fully downloaded, Media-Provider wil zip them per volume, into `.cbz` files with the following layout `Title Vol. ? Ch. ?.cbz`.


## Config

Mangadex supports a lot of different filtering options, the following are currently support.

### Include & Exclude tags (`includeTags`, `excludeTags`)
You may configure tags, the filter method is OR, so the manga will match at minimum one selected tag.

Add them as a `search_modifier` with key  or , use the English name as key

Ex.
- Official Colored
- Girls' Love
- Isekai

### Status (`status`)
The following values are possible
- ongoing
- completed
- hiatus
- cancelled

### Content Rating (`contentRating`)
The following values are possible
- safe
- suggestive
- erotica
- pornographic

### Publication Demographic (`publicationDemographic`)
- shounen
- shoujo
- josei
- seinen