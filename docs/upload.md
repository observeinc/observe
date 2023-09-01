# upload

    observe upload prompt my-notes.md

The upload command allows you to upload documents into your Observe instance.
Documents can extend your Observe instance, for example by adding company
specific context to the o11y help tool.

When uploading a document, you specify a document type, such as 'prompt' for
additions to the o11y feature. Other types may be supported in the future.

Documents are identified by their filename, or by an ID. No two documents can
have the same filename. If you specify a new filename and an existing document
ID, that document will be replaced by the new name. If this would create a
conflict, the upload will fail. If you don't specify an ID, then an existing
document of the same name will be replaced, or if the same name doesn't exist,
it will be created.

"Upload" is different from "create" because an uploaded document has both the
document contents itself, which may not be structured (consider an image used
as a backdrop, for example) and the metadata about that document (which is what
the normal object API deals with.)

## Supported Document Formats

The formats supported by the upload command are listed below:

### prompt

The 'prompt' document must be a text or markdown file (CommonMark is the
supported markdown flavor.)

The 'prompt' document will be split into segments sized approximately 1-2 kB in
size, and indexed for use within the o11y help tool. For example, if you upload
a list of error messages and common causes, then when you ask about those error
messages, o11y will be able to answer about them.

There can be up to 500 prompt documents.

## Example

    observe upload prompt oncall-schedule.md

## Example

    observe upload prompt --as-filename oncall-schedule.md path/to/new/file.md

