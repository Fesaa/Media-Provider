# Setup

Before being able to downloading any content, you'll have to set up pages. 
A page will be a tab in your navigation bar that allow you to search at least one provider for content.

A search request can be customized with some options, these will also be set up in the page's config.

Media Provider, provides a default set of pages which include a page for (most) providers. If you have no pages setup, the option to load these will be given on the dashboard.
The default pages expect the following directories

- Manga
- Movies
- Series
- Anime

# Server settings

You may set:
- custom root dir
- base url (`https://example.com/mp`)
- log handler
- log level
- cache type (redis is supported)

And the following download options
- Max concurrent torrents (hard max of 10)
- Max concurrent images (mangadex, ...) (hard max of 5)

## Combining providers

Not all providers support the same options, one may have a sort that works with `comments`, while the other doesn't. Depending on the provider, a wrong option might return an error.
Read the providers documentation to be sure.

## Modifiers

Modifiers are the option you can set per page. Which allow you to further customize your search request. 
Read providers documentation for possibilities.

## Directories

You must provide a page with at least one directory. This directory will be used as the base directory to download the content into.
You may change this directory while searching for content. 

If more than one directory is provided in setting, a dropdown will be provided to you in the search form.