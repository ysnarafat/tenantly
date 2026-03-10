# GitHub Issue Templates

This directory contains standardized issue templates for the Tenantly project.

## 📋 Available Templates

### Standard Templates
- **`bug_report.md`** - For reporting bugs
- **`feature_request.md`** - For suggesting new features

### Admin Hierarchy Epic (4-Phase Implementation)

The admin hierarchy epic is broken into 4 sequential phases, each with its own issue template:

```
[EPIC] feat(admin): implement 4-level admin hierarchy and multi-tenancy
├── [Phase 1] feat(admin): Database & Data Models
│   └── Duration: 2-3 days | Status: Ready to Start
│   └── Blocks: Phase 2
│
├── [Phase 2] feat(admin): Backend API Implementation
│   └── Duration: 5-7 days | Depends on: Phase 1 ✅
│   └── Blocks: Phase 3
│
├── [Phase 3] feat(admin): Frontend Components & UI
│   └── Duration: 5-7 days | Depends on: Phase 2 ✅
│   └── Blocks: Phase 4
│
└── [Phase 4] feat(admin): Testing, Security & Deployment
    └── Duration: 3-5 days | Depends on: Phase 3 ✅
    └── Closes parent epic
```

---

## 🚀 How to Use These Templates

### Creating the Parent Epic

1. Go to GitHub Issues → New Issue
2. Select **"Admin Hierarchy [EPIC]"** template
3. This creates an overview issue describing the entire 4-phase project
4. **Do not start implementation yet** - just merge this issue

### Creating Child Issues (In Order)

#### Phase 1: Database & Data Models
1. Once parent epic is merged, create a new issue
2. Select **"[Phase 1] Database & Data Models"** template
3. This templates provides all database schema changes, migrations, and Go model updates
4. Work through all checklist items
5. Merge PR when complete

#### Phase 2: Backend API Implementation
1. After Phase 1 PRs are merged to develop, create Phase 2 issue
2. Select **"[Phase 2] Backend API Implementation"** template
3. Implement services, handlers, middleware, and API endpoints
4. Must pass all integration tests before merge

#### Phase 3: Frontend Components & UI
1. After Phase 2 PRs are merged to develop, create Phase 3 issue
2. Select **"[Phase 3] Frontend Components & UI"** template
3. Build Angular components, services, and update existing components
4. E2E tests must pass

#### Phase 4: Testing, Security & Deployment
1. After Phase 3 PRs are merged to develop, create Phase 4 issue
2. Select **"[Phase 4] Testing, Security & Deployment"** template
3. Run full QA, security audit, database migration testing, and staged rollout
4. Close parent epic when complete

---

## 📐 Issue Structure

Each phase template includes:

1. **Overview** - What will be done and why
2. **Deliverables** - Specific tasks to complete
3. **Effort Estimate** - Duration and complexity
4. **Testing** - Unit and integration test requirements
5. **Acceptance Criteria** - Must-haves for completion
6. **Files Changed** - What files will be created/updated
7. **Related Issues** - Links to parent epic and blocking issues
8. **Progress Checklist** - Actionable task list

---

## 🔗 Template File Sizes

| Template | Size | Purpose |
|----------|------|---------|
| `admin-hierarchy.md` | 12K | **[PARENT EPIC]** Overview of 4-phase project |
| `phase-1-database.md` | 6.5K | Database tables, migrations, Go models |
| `phase-2-backend.md` | 9.9K | Services, handlers, middleware, API endpoints |
| `phase-3-frontend.md` | 13K | Angular components, services, UI |
| `phase-4-testing.md` | 13K | QA, security, migration, deployment |
| `bug_report.md` | 1.3K | Standard bug reporting |
| `feature_request.md` | 1.5K | Feature/enhancement suggestions |

---

## 💡 Best Practices

### When Creating Issues

1. **Use templates** - Never create blank issues (exceptions: urgent hotfixes)
2. **Fill completely** - Answer all fields, don't leave TODO items
3. **Add labels** - Use GitHub labels for quick filtering
4. **Link related issues** - Use "Blocks", "Depends on" relationships
5. **Estimate effort** - Help with sprint planning

### For The Admin Hierarchy Epic

1. **Start with parent** - Creates overview and gets team alignment
2. **One phase at a time** - Create next phase issue only after previous is merged
3. **Use checklists** - Check off items as you complete them
4. **Track progress** - Update issue status regularly
5. **Document blockers** - If stuck, add comment with details

---

## 🔄 Workflow Example

**Day 1**: Create parent epic (admin-hierarchy.md)
- Team reviews overview and asks questions
- Link to ADMIN_HIERARCHY_PLAN.md
- Merge when team is aligned

**Day 2**: Create Phase 1 issue (phase-1-database.md)
- Assign to backend engineer
- Create branch for migrations
- Start work immediately

**Day 5**: Phase 1 PR merged → Create Phase 2 issue
- Phase 2 depends on Phase 1 being complete
- Assign to backend engineer
- Implement services and API endpoints

**Day 10**: Phase 2 PR merged → Create Phase 3 issue
- Assign to frontend engineer
- Build components and UI

**Day 15**: Phase 3 PR merged → Create Phase 4 issue
- Assign to QA/DevOps engineer
- Run tests, security audit, staging rollout

**Day 20**: Phase 4 PR merged → Close parent epic ✅

---

## 📝 Configuration

GitHub automatically shows these templates when users click "New Issue":

```
.github/ISSUE_TEMPLATE/
├── admin-hierarchy.md          (shows in dropdown)
├── phase-1-database.md         (shows in dropdown)
├── phase-2-backend.md          (shows in dropdown)
├── phase-3-frontend.md         (shows in dropdown)
├── phase-4-testing.md          (shows in dropdown)
├── bug_report.md               (shows in dropdown)
├── feature_request.md          (shows in dropdown)
├── config.yml                  (configures template behavior)
└── README.md                   (this file)
```

See `config.yml` for links to documentation and security reporting.

---

## 🎯 Quick Reference

**Creating an issue:**
1. Go to GitHub Issues
2. Click "New Issue"
3. Select template from dropdown
4. Fill in all required fields
5. Submit issue

**For the admin hierarchy epic:**
1. Create parent: `admin-hierarchy.md`
2. Create child 1: `phase-1-database.md` (after parent merged)
3. Create child 2: `phase-2-backend.md` (after phase 1 merged)
4. Create child 3: `phase-3-frontend.md` (after phase 2 merged)
5. Create child 4: `phase-4-testing.md` (after phase 3 merged)
6. Close parent when phase 4 complete

---

## 📚 Related Documentation

- **Full Admin Hierarchy Plan**: See `ADMIN_HIERARCHY_PLAN.md` in repository root
- **Contributing Guide**: See `CONTRIBUTING.md` for code style and conventions
- **Authentication Details**: See `src/backend/api/AUTHENTICATION.md` for JWT implementation
- **Backend Architecture**: See `src/backend/api/README.md` for code structure

---

**Last Updated**: March 3, 2026
**Version**: 1.0
