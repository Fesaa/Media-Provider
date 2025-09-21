# Preferences

Each user can customize some of Media-Providers' behaviour. This can be done in the preference tab in settings. These
are scoped per user.

## General
![preferences_general.png](preferences_general.png)

Sometimes a matching cover can't be found, you can decide to use the first cover (volume/chapter 1), the latest last known
cover, or no cover at all when this happens. 

## Mappings

Not all providers offer the same quality of metadata, you may configure some mappings to filter and transform some of
the metadata to better suite your wants and needs.

<tabs>
<tab title="Blacklist">
Any tags, or genre's matching a value in here will be filtered out before being used to write ComicInfo.

![tags_blacklist.png](tags_blacklist.png)
</tab>
<tab title="Whitelist">
By default some providers will not save any tags, because there are so many that might not make sense. You may
configure some tags that will get saved. See provider documentation for details.

![tags_whitelist.png](tags_whitelist.png)
</tab>
<tab title="Genres">
Configure tags to be saved as genre for providers that don't make the distinction themselves.

![genres.png](genres.png)
</tab>
<tab title="Age rating mappings">
Map the presence of specific genre's or tags to an age rating. When more than one matches, the highest value is
used.

![age_ratings.png](age_ratings.png)
</tab>
<tab title="Tag mappings">
Transform tags before they're used. Use this to normalize different varicoses of the same tag. Matching happens
with normalization, but they may still be required if some providers use a different word, or plural for the same
genre/tag.

![tag_mappings.png](tag_mappings.png)
</tab>
</tabs>