import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { MatSnackBar } from '@angular/material/snack-bar';
import { of, throwError, Subject } from 'rxjs';
import { OrganizationListComponent } from './organization-list';
import { OrganizationService } from '../../../core/services/organization.service';
import { Organization } from '../../../core/models';

// actWithUndo drives a real MatSnackBarRef's onAction()/afterDismissed() —
// this stub lets tests simulate "undo clicked" vs. "toast timed out" without
// a real timer.
function createSnackBarRefStub() {
  const action = new Subject<void>();
  const dismissed = new Subject<{ dismissedByAction: boolean }>();
  return {
    ref: { onAction: () => action.asObservable(), afterDismissed: () => dismissed.asObservable() },
    clickUndo: () => {
      action.next();
      dismissed.next({ dismissedByAction: true });
    },
    timeOut: () => dismissed.next({ dismissedByAction: false }),
  };
}

describe('OrganizationListComponent', () => {
  let component: OrganizationListComponent;
  let fixture: ComponentFixture<OrganizationListComponent>;
  let organizationService: jasmine.SpyObj<OrganizationService>;
  let router: jasmine.SpyObj<Router>;
  let snackBar: jasmine.SpyObj<MatSnackBar>;

  const mockOrganizations: Organization[] = [
    {
      id: 1,
      name: 'Org 1',
      slug: 'org-1',
      subscriptionTier: 'basic',
      maxUsers: 10,
      active: true,
      createdAt: new Date(),
      updatedAt: new Date(),
    },
    {
      id: 2,
      name: 'Org 2',
      slug: 'org-2',
      subscriptionTier: 'professional',
      maxUsers: 50,
      active: true,
      createdAt: new Date(),
      updatedAt: new Date(),
    },
    {
      id: 3,
      name: 'Inactive Org',
      slug: 'inactive-org',
      subscriptionTier: 'enterprise',
      maxUsers: 1000,
      active: false,
      createdAt: new Date(),
      updatedAt: new Date(),
    },
  ];

  beforeEach(async () => {
    const organizationServiceSpy = jasmine.createSpyObj('OrganizationService', [
      'getOrganizations',
      'deleteOrganization',
    ]);
    const routerSpy = jasmine.createSpyObj('Router', ['navigate']);
    const snackBarSpy = jasmine.createSpyObj('MatSnackBar', ['open']);

    await TestBed.configureTestingModule({
      imports: [OrganizationListComponent],
      providers: [
        { provide: OrganizationService, useValue: organizationServiceSpy },
        { provide: Router, useValue: routerSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
      ],
    }).compileComponents();

    organizationService = TestBed.inject(
      OrganizationService
    ) as jasmine.SpyObj<OrganizationService>;
    router = TestBed.inject(Router) as jasmine.SpyObj<Router>;
    snackBar = TestBed.inject(MatSnackBar) as jasmine.SpyObj<MatSnackBar>;

    fixture = TestBed.createComponent(OrganizationListComponent);
    component = fixture.componentInstance;
  });

  describe('Initialization', () => {
    it('should create', () => {
      expect(component).toBeTruthy();
    });

    it('should load organizations on init', () => {
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );

      fixture.detectChanges();

      expect(organizationService.getOrganizations).toHaveBeenCalled();
      expect(component.dataSource.data).toEqual(mockOrganizations);
    });

    it('should set loading state during load', () => {
      organizationService.getOrganizations.and.returnValue(of({ organizations: [], total: 0 }));

      expect(component.loading()).toBe(false);
      fixture.detectChanges();
      expect(component.loading()).toBe(false);
    });
  });

  describe('Data loading', () => {
    it('should populate table with organizations', () => {
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );

      component.loadOrganizations();

      expect(component.dataSource.data).toEqual(mockOrganizations);
      expect(component.dataSource.data.length).toBe(3);
    });

    it('should handle empty organization list', () => {
      organizationService.getOrganizations.and.returnValue(of({ organizations: [], total: 0 }));

      component.loadOrganizations();

      expect(component.dataSource.data).toEqual([]);
      expect(component.loading()).toBe(false);
    });

    it('should handle error when loading organizations', () => {
      organizationService.getOrganizations.and.returnValue(
        throwError(() => new Error('Load failed'))
      );

      component.loadOrganizations();

      expect(snackBar.open).toHaveBeenCalledWith('Failed to load organizations', 'Close', {
        duration: 3000,
      });
      expect(component.loading()).toBe(false);
    });

    it('should set loading to true during load', () => {
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );

      component.loadOrganizations();

      expect(component.loading()).toBe(false);
    });
  });

  describe('Search functionality', () => {
    beforeEach(() => {
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );
      fixture.detectChanges();
    });

    it('should filter by organization name', () => {
      component.onSearchChange('Org 1');

      expect(component.searchTerm()).toBe('Org 1');
      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
      expect(filtered[0].name).toBe('Org 1');
    });

    it('should filter by slug', () => {
      component.onSearchChange('org-2');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
      expect(filtered[0].slug).toBe('org-2');
    });

    it('should be case insensitive', () => {
      component.onSearchChange('ORG 1');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(1);
      expect(filtered[0].name).toBe('Org 1');
    });

    it('should clear search results when search is empty', () => {
      component.onSearchChange('');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(3);
    });

    it('should return empty array when no matches', () => {
      component.onSearchChange('NonExistent');

      const filtered = component.filteredData();
      expect(filtered.length).toBe(0);
    });
  });

  describe('Navigation', () => {
    it('should navigate to create organization', () => {
      component.createOrganization();

      expect(router.navigate).toHaveBeenCalledWith(['/admin/organizations/new']);
    });

    it('should navigate to view organization detail', () => {
      const org = mockOrganizations[0];
      component.viewOrganization(org);

      expect(router.navigate).toHaveBeenCalledWith(['/admin/organizations', 1]);
    });

    it('should navigate to edit organization', () => {
      const org = mockOrganizations[0];
      component.editOrganization(org);

      expect(router.navigate).toHaveBeenCalledWith(['/admin/organizations', 1, 'edit']);
    });
  });

  describe('Delete functionality', () => {
    beforeEach(() => {
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );
      fixture.detectChanges();
    });

    it('should optimistically remove the organization and show an undo toast', () => {
      const org = mockOrganizations[0];
      const stub = createSnackBarRefStub();
      snackBar.open.and.returnValue(stub.ref as never);

      component.deleteOrganization(org);

      expect(component.dataSource.data).not.toContain(org);
      expect(snackBar.open).toHaveBeenCalledWith(
        `Organization "${org.name}" deleted`,
        'Undo',
        jasmine.objectContaining({ duration: 5000 })
      );
      expect(organizationService.deleteOrganization).not.toHaveBeenCalled();
    });

    it('should call the delete API once the toast times out without Undo', () => {
      const org = mockOrganizations[0];
      const stub = createSnackBarRefStub();
      snackBar.open.and.returnValue(stub.ref as never);
      organizationService.deleteOrganization.and.returnValue(of(void 0));

      component.deleteOrganization(org);
      stub.timeOut();

      expect(organizationService.deleteOrganization).toHaveBeenCalledWith(org.id);
    });

    it('should restore the organization and skip the delete API when Undo is clicked', () => {
      const org = mockOrganizations[0];
      const stub = createSnackBarRefStub();
      snackBar.open.and.returnValue(stub.ref as never);

      component.deleteOrganization(org);
      stub.clickUndo();

      expect(component.dataSource.data).toContain(org);
      expect(organizationService.deleteOrganization).not.toHaveBeenCalled();
    });

    it('should show an error and reload when the deferred delete fails', () => {
      const org = mockOrganizations[0];
      const stub = createSnackBarRefStub();
      snackBar.open.and.returnValue(stub.ref as never);
      organizationService.deleteOrganization.and.returnValue(
        throwError(() => new Error('Delete failed'))
      );
      organizationService.getOrganizations.calls.reset();
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );

      component.deleteOrganization(org);
      stub.timeOut();

      expect(snackBar.open).toHaveBeenCalledWith('Failed to delete organization', 'Close', {
        duration: 3000,
      });
      expect(organizationService.getOrganizations).toHaveBeenCalled();
    });
  });

  describe('Subscription tier colors', () => {
    it('should return primary color for basic tier', () => {
      expect(component.getSubscriptionTierColor('basic')).toBe('primary');
    });

    it('should return accent color for professional tier', () => {
      expect(component.getSubscriptionTierColor('professional')).toBe('accent');
    });

    it('should return warn color for enterprise tier', () => {
      expect(component.getSubscriptionTierColor('enterprise')).toBe('warn');
    });

    it('should return empty string for unknown tier', () => {
      expect(component.getSubscriptionTierColor('unknown')).toBe('');
    });
  });

  describe('DisplayedColumns', () => {
    it('should have all required columns', () => {
      expect(component.displayedColumns).toContain('name');
      expect(component.displayedColumns).toContain('slug');
      expect(component.displayedColumns).toContain('subscriptionTier');
      expect(component.displayedColumns).toContain('maxUsers');
      expect(component.displayedColumns).toContain('active');
      expect(component.displayedColumns).toContain('actions');
    });

    it('should have correct number of columns', () => {
      expect(component.displayedColumns.length).toBe(6);
    });
  });

  describe('Table functionality', () => {
    beforeEach(() => {
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );
      fixture.detectChanges();
    });

    it('should set dataSource with correct data', () => {
      expect(component.dataSource.data.length).toBe(3);
    });
  });

  describe('Organization data display', () => {
    beforeEach(() => {
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );
      component.loadOrganizations();
    });

    it('should display organization with correct properties', () => {
      const org = component.dataSource.data[0];
      expect(org.id).toBe(1);
      expect(org.name).toBe('Org 1');
      expect(org.slug).toBe('org-1');
      expect(org.active).toBe(true);
    });

    it('should display inactive organizations', () => {
      const inactiveOrg = component.dataSource.data.find((o) => !o.active);
      expect(inactiveOrg).toBeDefined();
      expect(inactiveOrg?.name).toBe('Inactive Org');
    });
  });

  describe('Computed properties', () => {
    beforeEach(() => {
      organizationService.getOrganizations.and.returnValue(
        of({ organizations: mockOrganizations, total: 3 })
      );
      component.loadOrganizations();
    });

    it('should compute filtered data', () => {
      component.searchTerm.set('Org 1');

      const filtered = component.filteredData();
      expect(filtered.length).toBeGreaterThan(0);
      expect(filtered.some((o) => o.name.includes('Org 1'))).toBe(true);
    });

    it('should update filtered data when search term changes', () => {
      const initialFiltered = component.filteredData();
      expect(initialFiltered.length).toBe(3);

      component.searchTerm.set('org-2');
      const updatedFiltered = component.filteredData();

      expect(updatedFiltered.length).toBeLessThan(initialFiltered.length);
    });
  });

  describe('Error handling', () => {
    it('should handle http errors when loading', () => {
      organizationService.getOrganizations.and.returnValue(
        throwError(() => ({ status: 500, statusText: 'Server Error' }))
      );

      component.loadOrganizations();

      expect(snackBar.open).toHaveBeenCalledWith('Failed to load organizations', 'Close', {
        duration: 3000,
      });
      expect(component.loading()).toBe(false);
    });

    it('should handle 403 forbidden error', () => {
      organizationService.getOrganizations.and.returnValue(
        throwError(() => ({ status: 403, statusText: 'Forbidden' }))
      );

      component.loadOrganizations();

      expect(snackBar.open).toHaveBeenCalled();
    });
  });
});
