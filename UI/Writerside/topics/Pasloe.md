# Pasloe (Custom)

Pasloe is the name of the download client that handles all non-torrent providers. It's a lot
more involved as it includes metadata, and subscriptions. Pasloe assumes each piece of content you're downloading
is a series with chapters. These chapters will be packed in each their own <code>.cbz</code> file.

<note>
    A chapter is an Oneshot if it has no volume <control>AND</control> chapter marker.
</note>

## File system structure

Pasloe saves series under a specific format. Changing files names, and to a lesser extent location may confuse the software
and cause chapters to be re-downloaded. Pasloe does not yet read <code>ComicInfo.xml</code>.

Series will be downloaded as follows

<code-block lang="text">
- Spice And Wolf
    - Spice and Wolf Vol. 1
        - Spice and Wolf Ch. 1.cbz
        - Spice and Wolf Ch. 2.cbz
    - Spice and Wolf Vol. 2
        - Spice and Wolf Ch. 3.cbz
        - Spice and Wolf Ch. 4.cbz
    - Spice and Wolf Ch. 5.cbz
</code-block>

If the series is a subscription, chapters may be re-downloaded if they are assigned a volume at a later date.

<tip>
    Set the environment variable <code>DISABLE_VOLUME_DIRS</code> to <code>TRUE</code> to flatten the series folder.
    Volume markers will be included when enabled.
</tip>

## ComicInfo.xml

Content downloaded via pasloe (Manga's) will include a <code>ComicInfo.xml</code> at root, with parsed metadata from the Provider.
See <a href="Preferences.md">Preferences</a> on how to configure and manipulate this metadata.

<warning>
    Not all providers have the same amount of metadata support, quality and correctness are up to the upstream data.
</warning>

## Download speed

Each provider has at most one series being downloaded at any given time. This is not configurable, the amount of images/s
in parallel is, with a max off 5. You may queue up to 100 series for each provider at once.

<note>
    Media-Provider does its best to not hit rate limits in the first place. When they're hit under special circumstances, 
    it will sleep until the time has passed, or for at least 1 minute if no header was found.
</note>