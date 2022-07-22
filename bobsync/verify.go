package bobsync

// after reading bobfile, need to verify that:
// * collection root and all underneath is in .gitignore [could be allowed with a warning]
// * collection root is somewhere in bob workspace [could be optional]
// * name and version are not allowed to include the separator char

// after parsing the tree of a collection, need to verify that:
// * no symlinks included
