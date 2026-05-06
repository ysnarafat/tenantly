# i18n Translation Key Organization — Production Approaches

**Date:** 2026-05-03  
**Context:** Frontend i18n growing with Bangla translation effort. Current approach creates duplication and scalability issues as new features are added.

## Problem Statement

The project currently has ~208 translation keys across 12 top-level namespaces in two monolithic JSON files (`en.json`, `bn.json`). As features are added, identical keys are duplicated across namespaces:

| Key | Copies today | Grows to with full app |
|---|---|---|
| `BUTTONS.CANCEL` | 6 dialogs | ~12+ |
| `BUTTONS.CREATE` / `BUTTONS.UPDATE` | 4 forms | ~8+ |
| `ERRORS.REQUIRED` | 3 dialogs | ~8+ |
| `MENUS.EDIT` / `MENUS.DELETE` / `MENUS.VIEW_DETAILS` | 2 lists | ~4+ |
| `FILTERS.CLEAR` / `FILTERS.ALL` / `FILTERS.STATUS_LABEL` | 2 lists | ~4+ |
| `RESULTS.SINGULAR` / `RESULTS.PLURAL` | 2 lists | ~4+ |
| `EMPTY_STATE.TRY_ADJUST` | 2 lists | ~4+ |
| `TABLE.STATUS` / `TABLE.ACTIONS` / `TABLE.TYPE` | 2 lists | ~5+ |

Every new dialog adds `BUTTONS.CANCEL`, `BUTTONS.UPDATE`, `BUTTONS.CREATE` again. This creates:
1. **Maintenance burden:** Update "Cancel" in one place? Have to check 6+ files. Risk of inconsistency.
2. **Translator confusion:** Same word appears in multiple sections with different meanings (context).
3. **Key explosion:** 208 keys today → 400+ keys by feature completion.
4. **Translation file merge conflicts:** Team working on different features collides in single monolithic JSON.
5. **Scalability mismatch:** Components are lazy-loaded per route, but translations load all-at-once at startup.

---

## Current State

### Structure
- **Single file per language:** `assets/i18n/en.json` (376 lines), `assets/i18n/bn.json` (376 lines)
- **Naming convention:** `FEATURE.SECTION.KEY` (e.g., `LEASE_LIST.FILTERS.SEARCH_LABEL`)
- **Nesting depth:** Inconsistent — `DUE_LIST` uses 2 levels, others use 3 levels max
- **Loading strategy:** Eager at app startup via `provideTranslateHttpLoader`, all 208 keys fetched before routing
- **No COMMON namespace:** Every feature redefines identical keys

### Key Distribution

| Namespace | Keys |
|---|---|
| DUE_LIST | ~12 |
| LEASE_LIST | ~35 |
| TENANT_LIST | ~32 |
| DASHBOARD | ~18 |
| CREATE_LEASE_DIALOG | ~14 |
| EDIT_LEASE_DIALOG | ~12 |
| PROPERTY_FORM_DIALOG | ~16 |
| TENANT_FORM_DIALOG | ~15 |
| LOGIN | ~8 |
| BUILDING_FORM_DIALOG | ~18 |
| UNIT_FORM_DIALOG | ~18 |
| PROPERTY_LIST | ~10 |

---

## Three Production-Grade Approaches

### Approach A: COMMON Namespace (Recommended for now)

**What:** Add a `COMMON` top-level namespace for shared vocabulary. Feature namespaces keep feature-specific keys only.

**Structure:**
```json
{
  "COMMON": {
    "BUTTONS": {
      "CANCEL": "Cancel",
      "SAVE": "Save",
      "CREATE": "Create",
      "UPDATE": "Update",
      "DELETE": "Delete",
      "EDIT": "Edit",
      "CLOSE": "Close",
      "RETRY": "Retry"
    },
    "ERRORS": {
      "REQUIRED": "Required",
      "MIN_VALUE": "Must be at least {{min}}",
      "MAX_VALUE": "Must be at most {{max}}",
      "MAX_LENGTH": "Max length is {{max}}"
    },
    "STATUS": {
      "ACTIVE": "Active",
      "INACTIVE": "Inactive",
      "LOADING": "Loading..."
    },
    "ACTIONS": {
      "VIEW_DETAILS": "View Details",
      "MORE": "More actions",
      "CLEAR": "Clear"
    },
    "PAGINATION": {
      "RESULT": "result",
      "RESULTS": "results"
    },
    "EMPTY": {
      "TRY_ADJUST": "Try adjusting your search or filters.",
      "NO_ITEMS": "No items found"
    },
    "HINTS": {
      "OPTIONAL": "Optional"
    }
  },
  "LEASE_LIST": { /* feature-specific only */ },
  "TENANT_LIST": { /* feature-specific only */ }
}
```

**Template Usage:**
```html
<!-- Shared vocabulary -->
<button>{{ 'COMMON.BUTTONS.CANCEL' | translate }}</button>
<mat-error>{{ 'COMMON.ERRORS.REQUIRED' | translate }}</mat-error>

<!-- Feature-specific -->
<th>{{ 'LEASE_LIST.TABLE.MONTHLY_RENT' | translate }}</th>
```

**Pros:**
- ✅ Zero new dependencies
- ✅ Zero config changes
- ✅ Eliminates ~40% of current keys (63 fewer redundant keys)
- ✅ Single file still — no deployment complexity
- ✅ Works immediately in all templates
- ✅ Foundation for later migration to Approach B
- ✅ Prevents future duplication

**Cons:**
- ❌ Single monolithic file still grows over time (eventual limit: ~500-1000 keys)
- ❌ All translations loaded at startup regardless of route (minor payload impact, ~5-10KB gzipped)
- ❌ Requires discipline: risk of over-populating COMMON with edge-case keys

**Best for:** This project right now. Low effort, immediate benefit, low risk.

**Effort:** 3-4 hours — refactor `en.json` + `bn.json`, update templates in 6-8 feature components, verify build.

---

### Approach B: Feature-Split Files with Multi-Loader

**What:** Split translations per feature domain into separate files. Use `@ngx-translate/multi-http-loader` to merge them at startup.

**Structure:**
```
assets/i18n/
  en/
    shared.json           ← COMMON.* keys
    auth.json             ← LOGIN.*
    dashboard.json        ← DASHBOARD.*
    leases.json           ← LEASE_LIST.*, CREATE_LEASE_DIALOG.*, EDIT_LEASE_DIALOG.*
    properties.json       ← PROPERTY_LIST.*, PROPERTY_FORM_DIALOG.*, BUILDING_FORM_DIALOG.*, UNIT_FORM_DIALOG.*
    tenants.json          ← TENANT_LIST.*, TENANT_FORM_DIALOG.*
    payments.json         ← (future) PAYMENT_LIST.*
    admin.json            ← (future) USER_LIST.*, ORG_LIST.*, AUDIT_LOG.*
  bn/
    (mirrors en/)
```

**Setup Change:**
```typescript
// app.config.ts
import { MultiTranslateHttpLoader } from 'ngx-translate-multi-http-loader';

export function HttpLoaderFactory(http: HttpClient) {
  return new MultiTranslateHttpLoader(http, [
    { prefix: './assets/i18n/en/', suffix: '.json' },
    { prefix: './assets/i18n/en/', suffix: '.json' }
  ]);
}

export const appConfig: ApplicationConfig = {
  providers: [
    provideTranslateLoader(HttpLoaderFactory),
    // ... rest of config
  ]
};
```

**Pros:**
- ✅ Clear ownership per feature (team A owns `leases.json`, team B owns `properties.json`)
- ✅ Small diffs per feature — merge conflicts only in single feature file, not monolithic JSON
- ✅ Scales cleanly to 20+ feature modules
- ✅ Translator can work on one feature translation without touching others
- ✅ Natural fit for team-based development

**Cons:**
- ❌ Requires `npm install ngx-translate-multi-http-loader` (7-10KB additional)
- ❌ All files still loaded eagerly at startup (N HTTP requests instead of 1, but cached)
- ❌ Slightly more complex setup (multi-loader configuration)
- ❌ Developers must remember: new feature → create its translation file

**Best for:** Projects with 3+ developers or 10+ feature modules. Industry standard for mid-to-large Angular apps (used by Taiga UI, enterprise Angular codebases).

**Effort:** 6-8 hours — install multi-loader, split JSON into 6 files, verify merge logic, update provider, test.

**Migration path:** Easy from Approach A — key names don't change, only file organization.

---

### Approach C: Lazy-Loaded Translations per Feature Route

**What:** Each lazy-loaded feature route loads only its own translation file via a route resolver. Only the translations for the current route are in memory.

**Structure:**
```typescript
// features/leases/leases.routes.ts
export const LEASE_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./lease-list/lease-list'),
    resolve: {
      i18n: () => inject(TranslateService).getTranslation('leases')
    }
  }
];

// In route guard or resolver:
ngOnInit() {
  this.translateService.use('en', 'leases'); // load only leases.json
}
```

**Pros:**
- ✅ Minimal initial payload (only shared.json at startup, feature files on route navigation)
- ✅ Routes only load translations they need
- ✅ Scales to 50+ feature modules (Nx monorepo pattern)
- ✅ Best performance for large apps (Stripe, Google Cloud Console use this pattern)

**Cons:**
- ❌ **Flash of Untranslated Content (FOUC)** during route navigation — labels appear in `i18n-key` format briefly
- ❌ Significantly more complex: requires route resolvers + `TranslateModule.forChild()` with `isolate: true`
- ❌ Not the default ngx-translate path — requires deep knowledge of the library
- ❌ Higher risk of bugs (async timing issues, loading states)
- ❌ Testing complexity increases

**Best for:** Very large apps (10,000+ keys, enterprise Nx monorepos). Overkill for this project.

**Effort:** 10-12 hours — complex resolver design, route timing coordination, FOUC mitigation, extensive testing.

---

## Recommendation: A → B → C (Progressive)

**Immediately (Approach A):**
- Extract COMMON keys into shared namespace
- Update templates (6-8 files affected)
- Eliminates ~30% duplication
- Zero risk, immediate payoff
- Prepares foundation for Approach B

**When team grows or 15+ modules (Approach B):**
- Install `multi-http-loader`
- Split existing JSON into feature files
- Key names don't change — templates already reference `LEASE_LIST.TABLE.TENANT` + `COMMON.BUTTONS.CANCEL`, so no template updates needed
- Non-breaking migration

**When at enterprise scale (Approach C):**
- Only if bundle size or performance becomes critical
- Requires dedicated effort for resolver + FOUC mitigation

---

## COMMON Keys to Extract (Approach A)

### BUTTONS
- `CANCEL` → "Cancel" (appears 6× today)
- `SAVE` → "Save"
- `CREATE` → "Create" (appears 4× today)
- `UPDATE` → "Update" (appears 4× today)
- `DELETE` → "Delete" (appears 2× today)
- `EDIT` → "Edit" (appears 2× today)
- `CLOSE` → "Close"
- `RETRY` → "Retry"
- `ADD` → "Add"

### ERRORS
- `REQUIRED` → "Required" (appears 3× today)
- `MIN_VALUE` → "Minimum value is {{min}}"
- `MAX_VALUE` → "Maximum value is {{max}}"
- `MAX_LENGTH` → "Maximum length is {{max}}"
- `MIN_LENGTH` → "Minimum length is {{min}}"
- `PATTERN` → "Invalid format"

### STATUS
- `ACTIVE` → "Active"
- `INACTIVE` → "Inactive"
- `LOADING` → "Loading..."

### ACTIONS (List/Table interactions)
- `VIEW_DETAILS` → "View Details" (appears 2× today)
- `MORE` → "More actions" (appears 2× today)
- `CLEAR` → "Clear" (appears 2× today)

### PAGINATION
- `RESULT` → "result" (appears 2× today)
- `RESULTS` → "results" (appears 2× today)

### EMPTY_STATE
- `TRY_ADJUST` → "Try adjusting your search or filters." (appears 2× today)
- `NO_ITEMS` → "No items found"

### HINTS
- `OPTIONAL` → "Optional"

---

## Bangla Translations for COMMON (Extract from existing)

```json
{
  "COMMON": {
    "BUTTONS": {
      "CANCEL": "বাতিল",
      "SAVE": "সংরক্ষণ করুন",
      "CREATE": "তৈরি করুন",
      "UPDATE": "আপডেট করুন",
      "DELETE": "মুছুন",
      "EDIT": "সম্পাদনা",
      "CLOSE": "বন্ধ করুন",
      "RETRY": "পুনরায় চেষ্টা করুন",
      "ADD": "যোগ করুন"
    },
    "ERRORS": {
      "REQUIRED": "প্রয়োজনীয়",
      "MIN_VALUE": "ন্যূনতম মান {{min}}",
      "MAX_VALUE": "সর্বাধিক মান {{max}}",
      "MAX_LENGTH": "সর্বাধিক দৈর্ঘ্য {{max}}",
      "MIN_LENGTH": "ন্যূনতম দৈর্ঘ্য {{min}}",
      "PATTERN": "অবৈধ বিন্যাস"
    },
    "STATUS": {
      "ACTIVE": "সক্রিয়",
      "INACTIVE": "নিষ্ক্রিয়",
      "LOADING": "লোড হচ্ছে..."
    },
    "ACTIONS": {
      "VIEW_DETAILS": "বিস্তারিত দেখুন",
      "MORE": "আরও বিকল্প",
      "CLEAR": "সাফ করুন"
    },
    "PAGINATION": {
      "RESULT": "ফলাফল",
      "RESULTS": "ফলাফল"
    },
    "EMPTY_STATE": {
      "TRY_ADJUST": "আপনার অনুসন্ধান বা ফিল্টার সামঞ্জস্য করার চেষ্টা করুন।",
      "NO_ITEMS": "কোনো আইটেম পাওয়া যায়নি"
    },
    "HINTS": {
      "OPTIONAL": "ঐচ্ছিক"
    }
  }
}
```

---

## Implementation Checklist (Approach A)

### Phase 1: Refactor JSON Files

- [ ] Add `COMMON` section to `en.json` with keys listed above
- [ ] Add `COMMON` section to `bn.json` with Bangla translations
- [ ] Remove duplicate `BUTTONS.CANCEL` from: CREATE_LEASE_DIALOG, EDIT_LEASE_DIALOG, PROPERTY_FORM_DIALOG, BUILDING_FORM_DIALOG, UNIT_FORM_DIALOG, TENANT_FORM_DIALOG
- [ ] Remove duplicate `BUTTONS.CREATE`/`UPDATE` from all form dialogs
- [ ] Remove duplicate `ERRORS.REQUIRED` from EDIT_LEASE_DIALOG, BUILDING_FORM_DIALOG, UNIT_FORM_DIALOG
- [ ] Remove duplicate `MENUS.*` from LEASE_LIST, TENANT_LIST (move to COMMON.BUTTONS.EDIT/DELETE, COMMON.ACTIONS.VIEW_DETAILS)
- [ ] Remove duplicate `FILTERS.CLEAR`, `FILTERS.ALL`, `FILTERS.STATUS_LABEL` from LEASE_LIST, TENANT_LIST
- [ ] Remove duplicate `RESULTS.*` from LEASE_LIST, TENANT_LIST
- [ ] Remove duplicate `EMPTY_STATE.TRY_ADJUST` from LEASE_LIST, TENANT_LIST
- [ ] Verify JSON syntax: `npm run build` passes with no missing key errors

### Phase 2: Update Templates (6-8 files)

For each affected template, replace feature-scoped keys with `COMMON` equivalents:

**Files to update:**
- `lease-list.html` → Replace `LEASE_LIST.MENUS.*`, `LEASE_LIST.FILTERS.CLEAR`, `LEASE_LIST.RESULTS.*`, `LEASE_LIST.EMPTY_STATE.TRY_ADJUST`
- `tenant-list.html` → Same as lease-list
- `create-lease-dialog.html` → Replace `CREATE_LEASE_DIALOG.BUTTONS.*`
- `edit-lease-dialog.html` → Replace `EDIT_LEASE_DIALOG.BUTTONS.*`, `EDIT_LEASE_DIALOG.ERRORS.REQUIRED`
- `property-form-dialog.html` → Replace `PROPERTY_FORM_DIALOG.BUTTONS.*`
- `building-form-dialog.html` → Replace `BUILDING_FORM_DIALOG.BUTTONS.*`, `BUILDING_FORM_DIALOG.ERRORS.*`
- `unit-form-dialog.html` → Replace `UNIT_FORM_DIALOG.BUTTONS.*`, `UNIT_FORM_DIALOG.ERRORS.*`
- `tenant-form-dialog.html` → Replace `TENANT_FORM_DIALOG.BUTTONS.*`

### Phase 3: Verify and Test

- [ ] `npm run build` passes without errors
- [ ] `npm run lint` passes
- [ ] `npm start:local` and manually toggle language (English/Bangla) on all updated screens
- [ ] Verify all labels show correctly in both languages
- [ ] Check for orphaned translation keys: `grep -r "LEASE_LIST.MENUS" src/frontend/src/app` should return only comments/docs
- [ ] Run `npm test` to ensure no broken specs

### Phase 4: Commit

- Commit message:
  ```
  refactor(i18n): extract common translation keys into COMMON namespace
  
  - Add COMMON section for shared vocabulary (BUTTONS, ERRORS, ACTIONS, etc)
  - Remove ~63 duplicate keys across features
  - Update templates in 8 components to reference COMMON.*
  - Reduces key count by 30% (~208 → ~145)
  - Prepares foundation for feature-split JSON files (Approach B) if needed later
  ```

---

## Parameterized Keys (Optional Enhancement)

For strings that differ only by entity name, use `ngx-translate` interpolation:

**Instead of:**
```json
{
  "PROPERTY_FORM_DIALOG": {
    "HINTS": {
      "PROPERTY_CODE": "Unique identifier for this property"
    }
  },
  "BUILDING_FORM_DIALOG": {
    "HINTS": {
      "BUILDING_CODE": "Unique identifier for this building"
    }
  }
}
```

**Use:**
```json
{
  "COMMON": {
    "HINTS": {
      "UNIQUE_ID": "Unique identifier for this {{entity}}"
    }
  }
}
```

**Template:**
```html
<mat-hint>{{ 'COMMON.HINTS.UNIQUE_ID' | translate: { entity: 'property' | translate } }}</mat-hint>
```

This further reduces key count but slightly increases template complexity.

---

## Expected Outcomes

After implementing Approach A:

| Metric | Before | After |
|---|---|---|
| Total keys (both locales) | ~416 (208×2) | ~290 (~145×2) |
| Duplication instances | 15+ | 0 |
| Files with BUTTONS.CANCEL | 6 | 1 (COMMON) |
| Feature namespaces with ERRORS.REQUIRED | 3 | 1 (COMMON) |
| Setup complexity | Low | Low (unchanged) |
| Maintenance burden | High | Medium (updates in one place) |

---

## Future Considerations

### When to migrate to Approach B:
- Team size grows to 3+ developers
- Feature count exceeds 15 modules
- Merge conflicts in JSON become frequent
- Need for per-feature translation ownership

### When to migrate to Approach C:
- Total keys exceed 10,000
- Initial page load performance critical
- Enterprise Nx monorepo structure in place

### Other enhancements to consider:
- Translation management platform integration (Crowdin, POEditor) — easier per-file management
- Automated key extraction from templates via `angular-localize` (alternative to ngx-translate)
- Translation coverage metrics — track % translated vs. % English fallback
- Missing key warnings in dev mode (detect `[i18n-key]` format in UI)

---

## References

- [ngx-translate/core documentation](https://github.com/ngx-translate/core)
- [ngx-translate/multi-http-loader](https://github.com/ngx-translate/multi-http-loader)
- [Angular i18n guide](https://angular.io/guide/i18n)
- Production patterns in: Taiga UI, Clarity Design, NgRx demo apps, enterprise Angular codebases
