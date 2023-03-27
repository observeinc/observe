# get

Ask the Observe API for the data of a specific instance of some object.

You can see the kinds of objects supported by running:

    observe help objects

You need to specify both the object type and the object id. For example:

    observe get workspace 41042069

The output will be in YAML format, with the top section containing the object
type and ID, and sections for "config" and "state". Config is properties that
you can change directly about an object; state is properties that are somehow
derived from other configurations or the system itself.

For example, when you save a dataset, the "name" is a config property, whereas
the modification date is a state property, because it is derived by the system
when saving, rather than provided as direct input.
