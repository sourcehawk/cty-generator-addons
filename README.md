# CTY Generator Addons

Small add-on for cty that parses custom doc tags from your Go CRD types and generates HTML for it on the CTY style. 

> [!CAUTION]️ Note on authorship: many parts of this repository were written with the assistance of AI and have not been
> tested. Please review carefully and treat this as community-supported. 

section with Conditions and 
their Reasons, then injects it into the CTY-generated docs.

## Condition API Addon

### What it does

Scans your Go sources for special tags in // comments:

- Tag to declare a constant, field or type as a condition of a CRD: `// +cty:condition:for=<CRD>`
- Tag to declare a constant or type as a condition of a CRD: `// +cty:reason:for=<CRD>/<Condition>`

Renders a collapsible “Conditions” section per CRD. Injects the generated HTML at the end of the 
`class="content"` block in the CTY HTML

### How to use the tags

```go
// api/v1alpha1/conditions.go
package v1alpha1

// The type that holds your condition constants
type ZeebeClusterConditionType string

const (
	// ZeebeClusterReadyCondition The condition indicates whether the ZeebeCluster is ready
	//  - true condition status means the cluster is healthy
	//  - false condition status means the cluster is not healthy
	//  - unknown condition status with a reason means the cluster is in long transition (starting, updating, etc.)
	// +cty:condition:for=ZeebeCluster
	ZeebeClusterReadyCondition ZeebeClusterConditionType = "Ready"

	// EncryptionReadyCondition The condition indicates whether the cluster's encryption is ready
	//  - status = true: The encryption is ready
	//  - status = false: The encryption is not ready
	// +cty:condition:for=ZeebeCluster
	EncryptionReadyCondition ZeebeClusterConditionType = "EncryptionReady"
)

// A type that groups your reasons for a specific condition
type EncryptionReadyReason string

const (
	// ExternalEncryptionKeyNotSupplied is surfaced when external encryption is configured on cluster creation.
	// When active, cluster progression is halted while waiting for the user to supply their external key ID.
	// +cty:reason:for=ZeebeCluster/EncryptionReady
	ExternalEncryptionKeyNotSupplied EncryptionReadyReason = "ExternalEncryptionKeyNotSupplied"

	// EncryptionReady indicates that the encryption is ready and active.
	// +cty:reason:for=ZeebeCluster/EncryptionReady
	EncryptionReady EncryptionReadyReason = "Ready"

	// ExternalEncryptionKeyNotReady indicates that the external encryption key is not ready yet.
	// +cty:reason:for=ZeebeCluster/EncryptionReady
	ExternalEncryptionKeyNotReady EncryptionReadyReason = "ExternalEncryptionKeyNotReady"

	// EncryptedStorageNotReady indicates that the encrypted storage is not ready yet.
	// +cty:reason:for=ZeebeCluster/EncryptionReady
	EncryptedStorageNotReady EncryptionReadyReason = "EncryptedStorageNotReady"

	// EncryptionCreationError indicates that an error occurred setting up the encryption.
	// +cty:reason:for=ZeebeCluster/EncryptionReady
	EncryptionCreationError EncryptionReadyReason = "CreationError"
)
```

### Output

![conditions](docs/conditions_generator.png)

### How to install the binary

Install the binary with 

```bash
go install github.com/sourcehawk/cty-generator-addons/cmd/cty-conditions-addon@latest
```

### How to run

First generate the cty HTML as usual

```bash
cty generate crd \
  --folder config/crd/bases/ \
  --format html \
  --output docs/api/index.html
```

Then run the condition addon binary to inject the additional HTML into the cty index.html

```bash
cty-conditions-addon \
  -path ./api \
  -title "API Conditions" \
  -inject-into ./docs/api/index.html
```


