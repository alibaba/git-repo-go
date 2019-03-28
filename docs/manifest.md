# Manifest parsing

Manifest structure is one of the build blocks for `git-repo`.  Parsing of
manifest is implemented in `manifest/manifest.go`.


# Load manifest

Use the following structures for parsing manifest XML file.

    Manifest      // toplevel xml element
    Remote        // for remote XML element
    Default       // for default XML element
    Server        // for manifest-server element
    Project       // for project XML element
    Annotation    // for annotation XML element
    CopyFile      // for copy-file XML element
    LinkFile      // for link-file XML element
    ExtendProject // for extend-project XML element
    RemoveProject // for remove-project XML element
    RepoHooks     // for repo-hook XML element
    Include       // for include XML element

The Load() function will read the XML file, and use "encoding/xml" package
to unmarshal content of the XML file.


# Merge of manifests

One manifest XML can include other XML files using "include" directive.

    <include name="other-manifest.xml"></include>

And user defined manifest XML files (XML files in `.repo/local-manifests`
directory) are also merged with repo manifest.

The `Merge()` function of Manifest object helps to merge manifests.


# Testing

To test manifest manipulation, test cases are added in file
`manifest/manifest_test.go`.
