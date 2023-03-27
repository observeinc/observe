# list

Ask the Observe API to list any instances of various kinds of objects.

The list will show the ID, name, and perhaps a few other important properties
of each object instance. To see the full information about an object, consider
`observe get objectype objectid`

You can see the kinds of objects supported by running:

    observe help objects

To list only objects with ID or name matching some particular substring, use

    observe list <objecttype> <substring>

## Example

    observe list dataset Log

    observe list workspace
