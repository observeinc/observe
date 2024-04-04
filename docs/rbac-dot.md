# rbac-dot

Plot the relationships between users, groups, and objects.

This command lets you specify a user, a group, or an object, and will plot
a graph of the relationships between them.

The "plotting" means outputting a GraphViz DOT file to standard output, which
you then need to run through the "dot" command to generate a PNG or SVG.

## Example

    observe rbac-dot --user 12345
