package shared

import (
	"os"
)

// LX188MetaOnly reports whether the metadata-only migration spike is enabled.
//
// PoC ONLY, NOT FOR MERGE. This is a stand-in for a negotiated migration mode so the
// feasibility question can be answered without plumbing a protocol change first. Set
// LXD_LX188_METAONLY=1 on both the source and the target daemon. Grep for LX188 to find
// every call site.
func LX188MetaOnly() bool {
	return os.Getenv("LXD_LX188_METAONLY") == "1"
}

// LX188Standby reports whether this daemon holds replicas rather than primaries.
//
// PoC ONLY, NOT FOR MERGE. This stands in for a project or pool level "this is a replica" flag.
// It is deliberately separate from the metadata-only switch, because the leader must keep writing
// backup files while the standby cannot write anything at all. Set LXD_LX188_STANDBY=1 on the
// standby only.
func LX188Standby() bool {
	return os.Getenv("LXD_LX188_STANDBY") == "1"
}
