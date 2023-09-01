# delete

Ask the Observe API to delete a particular object by id.

The command takes an object type, and an ID of that object. If you have
permission to delete that object, it will be deleted, and success will be
returned.

You can see the kinds of objects supported by running:

    observe help objects

## Example

    observe delete document o::1234567890:document:8007654321

