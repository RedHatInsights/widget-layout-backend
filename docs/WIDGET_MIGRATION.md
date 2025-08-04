# Widget Migration Guide

This document provides guidance for migrating widgets to the new FEO (Frontend Optimization) system with updated widget identifiers and CSS selectors.

## Widget ID Migration

When widgets are collected via the FEO system, their identifiers change from simple Chrome service IDs to compound identifiers that include both scope and module information. The following table shows the mapping between old and new widget identifiers:

| Chrome Service ID       | FEO Widget ID                                          |
|-------------------------|--------------------------------------------------------|
| `acs`                   | `landing-./AcsWidget`                                  |
| `ansible`               | `landing-./AnsibleWidget`                              |
| `edge`                  | `landing-./EdgeWidget`                                 |
| `exploreCapabilities`   | `landing-./ExploreCapabilities`                        |
| `favoriteServices`      | `chrome-./DashboardFavorites`                          |
| `imageBuilder`          | `landing-./ImageBuilderWidget`                         |
| `integrations`          | `sources-./IntegrationsWidget`                         |
| `learningResources`     | `learningResources-./BookmarkedLearningResourcesWidget` |
| `notificationsEvents`   | `notifications-./DashboardWidget`                      |
| `openshift`             | `landing-./OpenShiftWidget`                            |
| `openshiftAi`           | `landing-./OpenShiftAiWidget`                          |
| `quay`                  | `landing-./QuayWidget`                                 |
| `recentlyVisited`       | `landing-./RecentlyVisited`                            |
| `rhel`                  | `landing-./RhelWidget`                                 |
| `subscriptions`         | `subscriptionInventory-./SubscriptionsWidget`          |
| `supportCases`          | `landing-./SupportCaseWidget`                          |

## CSS Selector Migration

Due to the change in widget IDs, CSS selectors must also be updated to target the new identifiers. The new CSS class names are generated from the widget's full FEO identifier.

### Selector Format

CSS selectors are now based on the complete widget identifier, including special characters: `scope-widgetId`. It is common that the scope string will be duplicated in your CSS selector. Thats because the widget ID includes the scope substring. When targeting widgets with special characters in their identifiers (such as `./`), you must escape these characters in your CSS.

### Example Migration

For the "Explore Capabilities" widget, the CSS selector migration should be done gradually by adding new selectors alongside existing ones:

**Step 1: Add new selectors alongside existing ones**
```scss
/* Combined selectors - both old and new target the same styles */
.landing-exploreCapabilities,
[class*="landing-landing-./ExploreCapabilities"] {
  /* Widget styles - no duplication needed */
  background-color: #f5f5f5;
  border: 1px solid #ddd;
  padding: 16px;
  border-radius: 8px;
}

/* Alternative: Using escaped class selector */
.landing-exploreCapabilities,
.landing-landing-\.\/ExploreCapabilities {
  /* Widget styles */
  background-color: #f5f5f5;
  border: 1px solid #ddd;
  padding: 16px;
  border-radius: 8px;
}
```

**Step 2: Remove old selectors after migration is complete**
```scss
/* Only keep the new selectors once migration is fully deployed */
[class*="landing-landing-./ExploreCapabilities"] {
  /* Widget styles */
}
```

### CSS Selector Best Practices

- **Gradual migration**: Add new selectors alongside existing ones, don't replace immediately
- **Comma-separated selectors**: Use comma-separated selectors to avoid duplicating styles (e.g., `.old-selector, .new-selector { /* styles */ }`)
- **Migration window**: Keep both old and new selectors active during the transition period
- **Attribute selectors**: Use `[class*="widget-id"]` for widgets with special characters in their identifiers
- **Escaped selectors**: When using class selectors, escape special characters (`.` becomes `\.`, `/` becomes `\/`)
- **Specificity**: Be aware that attribute selectors have the same specificity as class selectors
- **Testing**: Test both old and new selectors during the migration period
- **Cleanup**: Remove old selectors only after confirming the new system is fully deployed

### Common Special Characters

When working with FEO widget identifiers, you may encounter these special characters that require escaping in CSS:

| Character | Escaped Form | Usage in Widget IDs |
|-----------|--------------|---------------------|
| `.`       | `\.`         | Module path separator |
| `/`       | `\/`         | Path separator |

## Migration Checklist

When migrating widgets to the FEO system, follow these steps to ensure a smooth transition:

### Phase 1: Preparation
1. **Identify affected widgets** by reviewing the widget ID mapping table above
2. **Audit existing CSS** to find all selectors targeting the old widget IDs
3. **Plan the migration timeline** to coordinate with deployment schedules

### Phase 2: Add New Selectors
4. **Add new CSS selectors** alongside existing ones (do not replace yet)
5. **Use proper escaping** for special characters in widget identifiers
6. **Test both selector sets** to ensure styling works with old and new IDs
7. **Update widget references** in code to support both old and new identifiers

### Phase 3: Validation
8. **Test widget functionality** to ensure proper rendering and behavior
9. **Validate CSS styling** in both old and new systems
10. **Verify responsive behavior** across different screen sizes
11. **Check for conflicts** between old and new selectors

### Phase 4: Cleanup (After Full Migration)
12. **Remove old CSS selectors** once the new system is fully deployed
13. **Clean up old widget references** in code
14. **Update documentation** to reflect the new widget identifiers only
15. **Archive migration notes** for future reference

For more information about widget configuration and management, see [docs/CONFIGURATION.md](docs/CONFIGURATION.md).
