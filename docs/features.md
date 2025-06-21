# Feature (config)

Some features/settings are behind an env variables instead of a UI options, as do not make sense to change in flight. 


### Disable Volume Dirs (`DISABLE_VOLUME_DIRS`)
Kavita has a bug where it rescans series on each scan loop when they have subfolders until this has been resolved, we will not be adding the volume dir. See https://github.com/Kareadita/Kavita/issues/3557 for more context.

### DisableOneShotInFileName (`DISABLE_ONE_SHOT_IN_FILE_NAME`)
Prevents Media-Provider from adding `(One Shot)` in the file name