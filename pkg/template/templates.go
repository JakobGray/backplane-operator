package templates

import "embed"

//go:embed components/assisted-service
var AssistedServiceFS embed.FS

//go:embed components/cluster-lifecycle
var ClusterLifecycleFS embed.FS

//go:embed components/cluster-manager
var ClusterManagerFS embed.FS

//go:embed components/cluster-proxy-addon
var ClusterProxyAddonFS embed.FS

//go:embed components/console-mce
var ConsoleMCEFS embed.FS

//go:embed components/discovery-operator
var DiscoveryFS embed.FS

//go:embed components/hive-operator
var HiveFS embed.FS

//go:embed components/hypershift
var HyperShiftFS embed.FS

//go:embed components/managed-serviceaccount
var ManagedServiceAccountFS embed.FS

//go:embed components/server-foundation
var ServerFoundationFS embed.FS

//go:embed crds
var CRDFS embed.FS
