# Widget Layout Backend Configuration

This document explains how the Widget Layout Backend loads and processes its configuration for base templates and widget mappings.

## Configuration Flow

The Widget Layout Backend loads its configuration from Kubernetes ConfigMaps through environment variables. The configuration is processed at application startup and stored in runtime registries for fast access.

### Environment Variables

The application expects two main configuration environment variables:

- `BASE_LAYOUTS` - JSON string containing base widget dashboard templates
- `WIDGET_MAPPING` - JSON string containing widget module federation metadata

### Kubernetes Integration

The configuration is mounted from ConfigMaps as defined in the `clowdapp.yaml`:

```yaml
env:
  # FEO generated base layout config
  - name: BASE_LAYOUTS
    valueFrom:
      configMapKeyRef:
        name: ${FEO_BASE_LAYOUTS_CONFIGMAP}
        key: base-widget-dashboard-templates.json
  # FEO generated widget mapping config
  - name: WIDGET_MAPPING
    valueFrom:
      configMapKeyRef:
        name: ${FEO_WIDGET_MAPPING_CONFIGMAP}
        key: widget-registry.json
```

## Runtime Loading Process

### Initialization Flow

1. **Application Startup**: The Go `init()` functions in service packages execute
2. **Config Loading**: Environment variables are read and parsed as JSON
3. **Registry Population**: Parsed data is stored in in-memory registries
4. **Error Handling**: Invalid configurations cause fatal errors and service shutdown

### Base Templates Loading

**File**: `pkg/service/BaseLayoutTemplate.go`

```go
func init() {
    cfg := config.GetConfig()
    if err := LoadBaseTemplatesFromConfig(cfg.BaseWidgetDashboardTemplates); err != nil {
        logrus.Fatalln("Failed to parse base widget dashboard templates, shutting down the service", err)
    }
}
```

The `LoadBaseTemplatesFromConfig` function:
- Parses JSON array of base templates
- Validates template structure
- Stores templates in `BaseTemplateRegistry`
- Logs successful loading

### Widget Mappings Loading

**File**: `pkg/service/WidgetMapping.go`

```go
func init() {
    cfg := config.GetConfig()
    if err := LoadWidgetMappingsFromConfig(cfg.WidgetMappingConfig); err != nil {
        logrus.Fatalln("Failed to parse widget mappings, shutting down the service", err)
    }
}
```

The `LoadWidgetMappingsFromConfig` function:
- Parses JSON array of widget mappings
- Generates unique keys for each widget
- Stores mappings in `WidgetMappingRegistry`
- Logs successful loading

#### Widget Key Generation

Widget mappings are stored in the registry using a unique key generated from the widget metadata:

**Key Format**: `{scope}-{module}[-{importName}]`

**Algorithm**:
```go
func (wc *WidgetModuleFederationMetadata) GetWidgetKey() string {
    key := fmt.Sprintf("%s-%s", wc.Scope, wc.Module)
    if wc.ImportName != nil && *wc.ImportName != "" {
        key = fmt.Sprintf("%s-%s", key, *wc.ImportName)
    }
    return key
}
```

**Examples**:
- `insights-vulnerabilities-widget` (no importName)
- `insights-compliance-widget-ComplianceWidget` (with importName)
- `monitoring-alerts-widget-AlertsWidget` (with importName)

This ensures each widget has a unique identifier even when multiple widgets come from the same scope/module combination.

## Configuration Formats

### Base Widget Dashboard Templates

**JSON Structure**:
```json
[
  {
    "name": "insights-dashboard",
    "displayName": "Insights Dashboard",
    "templateConfig": {
      "sm": [
        {
          "w": 1,
          "h": 4,
          "maxH": 10,
          "minH": 1,
          "cx": 0,
          "cy": 0,
          "i": "insights-compliance",
          "static": false
        }
      ],
      "md": [...],
      "lg": [...],
      "xl": [...]
    }
  }
]
```

**Field Descriptions**:
- `name`: Unique identifier for the template
- `displayName`: Human-readable name
- `templateConfig`: Responsive layout configuration
  - `sm`, `md`, `lg`, `xl`: Breakpoint-specific widget layouts
  - `w`, `h`: Widget width and height
  - `maxH`, `minH`: Maximum and minimum height constraints
  - `cx`, `cy`: Widget coordinates (see coordinate system section)
  - `i`: Widget identifier/type
  - `static`: Whether widget is locked in position

### Widget Module Federation Metadata

**JSON Structure**:
```json
[
  {
    "scope": "insights",
    "module": "dashboard-widget",
    "importName": "DashboardWidget",
    "featureFlag": "enable-insights-dashboard",
    "config": {
      "title": "Insights Dashboard",
      "icon": "insights-icon",
      "permissions": ["read:insights"],
      "headerLink": {
        "name": "View Details",
        "href": "https://console.redhat.com/insights"
      }
    },
    "defaults": {
      "w": 2,
      "h": 3,
      "maxH": 6,
      "minH": 1
    }
  }
]
```

**Field Descriptions**:
- `scope`: Module federation scope
- `module`: Module federation module name
- `importName`: (Optional) Specific import name
- `featureFlag`: (Optional) Feature flag for conditional loading
- `config`: Widget configuration
  - `title`: Widget display title
  - `icon`: Widget icon identifier
  - `permissions`: (Optional) Required permissions array
  - `headerLink`: (Optional) Header link configuration
- `defaults`: Default widget dimensions

## Coordinate System: cx/cy vs x/y

### Why cx/cy is Used in Configuration

Configuration files use `cx` and `cy` (config x/y) instead of `x` and `y` due to a YAML parser limitation:

**Technical Issue**: The YAML parser treats the character "y" as a reserved word that translates to the boolean value `true`, causing unmarshaling errors when processing widget coordinates.

**Solution**: 
- **Configuration files**: Use `cx` and `cy` exclusively
- **REST API**: Use `x` and `y` exclusively  
- **Runtime conversion**: `cx`/`cy` is automatically converted to `x`/`y` during configuration loading

### Format Separation

**Configuration Files** (ConfigMaps, JSON config):
```json
{
  "w": 2,
  "h": 3,
  "cx": 0,  // Configuration uses cx
  "cy": 1,  // Configuration uses cy
  "i": "widget-id"
}
```

**REST API** (Requests/Responses):
```json
{
  "w": 2,
  "h": 3,
  "x": 0,   // API uses x
  "y": 1,   // API uses y
  "i": "widget-id"
}
```

### Runtime Conversion

The `WidgetItem.UnmarshalJSON()` method handles the conversion. The actual implementation includes comprehensive type checking and reflection-based field mapping:

```go
func (wi *WidgetItem) UnmarshalJSON(data []byte) error {
    var temp map[string]interface{}
    if err := json.Unmarshal(data, &temp); err != nil {
        return err
    }

    // use original values
    if temp["x"] != nil && temp["y"] != nil {
        x, ok := temp["x"].(float64)
        if ok {
            vi := int(x)
            wi.X = &vi
        }
        // ... error handling
        y, ok := temp["y"].(float64)
        if ok {
            vi := int(y)
            wi.Y = &vi
        }
        // ... error handling
    } else if temp["cx"] != nil && temp["cy"] != nil {
        cx, ok := temp["cx"].(float64)
        if ok {
            vi := int(cx)
            wi.X = &vi
        }
        // ... error handling
        cy, ok := temp["cy"].(float64)
        if ok {
            vi := int(cy)
            wi.Y = &vi
        }
        // ... error handling
    } else if temp["x"] == nil && temp["y"] == nil && temp["cx"] == nil && temp["cy"] == nil {
        return errors.New("WidgetItem must have either x/y or cx/cy attributes")
    }

    // Additional reflection-based field mapping for other properties...
    // (Full implementation includes comprehensive field mapping logic)
}
```

*Note: The above shows the core coordinate conversion logic. The actual implementation in `api/common.go` includes additional reflection-based field mapping for all widget properties.*

### Configuration Guidelines

- **Use `cx` and `cy`** ONLY in JSON/YAML configuration files
- **Use `x` and `y`** ONLY in REST API requests/responses and runtime objects
- **No mixing formats**: Configuration files must use `cx`/`cy`, REST API must use `x`/`y`
- **Automatic conversion**: `cx`/`cy` is converted to `x`/`y` during configuration loading
- **Validation ensures** at least one coordinate system is present in configuration

## Registry Access

### Base Templates Registry

```go
// Access base templates
template, exists := service.BaseTemplateRegistry.GetBase("template-name")

// Get all base templates
templates := service.BaseTemplateRegistry.GetAllBases()
```

### Widget Mappings Registry

```go
// Access widget mappings
mapping, exists := service.WidgetMappingRegistry.GetWidgetMapping("widget-key")

// Get all widget mappings
mappings := service.WidgetMappingRegistry.GetAllWidgetMappings()
```

## Error Handling

### Configuration Errors

- **Invalid JSON**: Service fails to start with fatal error
- **Missing Required Fields**: Validation errors logged, service continues
- **Empty Configuration**: Handled gracefully, empty registries created

### Runtime Errors

- **Missing Templates**: HTTP 404 responses
- **Invalid Coordinates**: Unmarshaling errors with descriptive messages
- **Registry Access**: Safe fallbacks to empty collections

## Best Practices

1. **Validate Configuration**: Test JSON structure before deployment
2. **Use cx/cy**: Always use `cx` and `cy` in configuration files
3. **Handle Optionals**: Mark optional fields appropriately
4. **Monitor Logs**: Check startup logs for configuration loading status
5. **Feature Flags**: Use feature flags for conditional widget loading

## Example Configuration Files

### Minimal Base Template
```json
[
  {
    "name": "minimal-dashboard",
    "displayName": "Minimal Dashboard",
    "templateConfig": {
      "sm": [
        {
          "w": 1,
          "h": 2,
          "maxH": 4,
          "minH": 1,
          "cx": 0,
          "cy": 0,
          "i": "simple-widget"
        }
      ],
      "md": [...],
      "lg": [...],
      "xl": [...]
    }
  }
]
```

### Minimal Widget Mapping
```json
[
  {
    "scope": "basic",
    "module": "simple-widget",
    "config": {
      "title": "Simple Widget",
      "icon": "widget-icon"
    },
    "defaults": {
      "w": 1,
      "h": 2,
      "maxH": 4,
      "minH": 1
    }
  }
]
```

## Local Development Setup

For local development, environment variables can be set through several methods:

### Using .env File

Create a `.env` file in the project root:

```bash
# .env file for local development
BASE_LAYOUTS='[{"name":"local-dashboard","displayName":"Local Dashboard","templateConfig":{"sm":[],"md":[],"lg":[],"xl":[]}}]'
WIDGET_MAPPING='[{"scope":"local","module":"test-widget","config":{"title":"Test Widget","icon":"test-icon"},"defaults":{"w":2,"h":2,"maxH":4,"minH":1}}]'
```

### Using Environment Variables

Set variables directly in your shell:

```bash
# Set widget mapping
export WIDGET_MAPPING='[
  {
    "scope": "insights",
    "module": "test-widget", 
    "config": {
      "title": "Test Widget",
      "icon": "test-icon"
    },
    "defaults": {
      "w": 2,
      "h": 2,
      "maxH": 4,
      "minH": 1
    }
  }
]'

# Set base layout
export BASE_LAYOUTS='[
  {
    "name": "test-template",
    "displayName": "Test Template",
    "templateConfig": {
      "sm": [],
      "md": [],
      "lg": [],
      "xl": []
    }
  }
]'
```

### Development vs Production

| Environment | Configuration Source | Format |
|-------------|---------------------|---------|
| **Local Development** | Environment variables or `.env` file | JSON strings |
| **Production/Kubernetes** | ConfigMaps mounted as environment variables | JSON strings |

**Important Notes**:
- **Use `cx`/`cy`** in configuration JSON (not `x`/`y`)
- **JSON must be valid** - invalid JSON causes service startup failure
- **Empty configurations** are handled gracefully (empty registries)

This configuration system provides a flexible, scalable way to manage widget layouts and mappings while handling the technical constraints of YAML parsing and Kubernetes ConfigMap integration.
