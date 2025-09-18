# Mangadex

Any series on Mangadex can be downloaded. However, only one series will be downloaded from Mangadex at any given time.
You may queue up to 100 series at once, they will automatically start once one finishes unless otherwise configured.

## Search options

<table >
    <tr>
        <td>Key</td>
        <td>Explanation</td>
        <td>Default value</td>
        <td>Allowed types</td>
    </tr>
    <tr>
        <td>SkipNotFoundTags</td>
        <td>Ignore tags that can't be matched to a mangadex id, otherwise return an error</td>
        <td>true</td>
        <td>switch</td>
    </tr>
    <tr>
        <td>includeTags</td>
        <td>Tags the series must match</td>
        <td>N/A</td>
        <td>multi</td>
    </tr>
    <tr>
        <td>includedTagsMode</td>
        <td>Match all <code>AND</code> or at least one <code>OR</code> of the tags</td>
        <td><code>AND</code></td>
        <td>dropdown</td>
    </tr>
    <tr>
        <td>excludeTags</td>
        <td>Tags the series must not match</td>
        <td>N/A</td>
        <td>multi</td>
    </tr>
    <tr>
        <td>excludeTagsMode</td>
        <td>Match none <code>AND</code> or at most one <code>OR</code> of the tags</td>
        <td><code>OR</code></td>
        <td>dropdown</td>
    </tr>
    <tr>
        <td>status</td>
        <td>Publication status, one of <code>ongoing</code>, <code>completed</code>, <code>hiatus</code>, <code>cancelled</code></td>
        <td>N/A</td>
        <td>multi</td>
    </tr>
    <tr>
        <td>contentRating</td>
        <td>Age rating, one of <code>safe</code>, <code>suggestive</code>, <code>erotica</code>, <code>pornographic</code></td>
        <td>N/A</td>
        <td>multi</td>
    </tr>
    <tr>
        <td>publicationDemographic</td>
        <td>Publication demographic, one of <code>josei</code>, <code>seinen</code>, <code>shoujo</code>, <code>shounen</code></td>
        <td>N/A</td>
        <td>multi</td>
    </tr>
</table>

## Download options

Download options are not configurable, and provided by the server. They'll be explained in the menu. 



<tip>
    Set a title override when download subscription to prevent upstream title changes from confusing Media-Provider
</tip>