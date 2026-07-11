import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { OverlayContainer } from '@angular/cdk/overlay';
import { MatMenuTrigger } from '@angular/material/menu';
import { of } from 'rxjs';
import { TranslateModule } from '@ngx-translate/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { TenantList } from './tenant-list';
import { TenantService } from '../../../core/services/tenant.service';
import { PermissionService } from '../../../core/services/permission.service';
import { Tenant, TenantListResponse } from '../../../core/models/tenant.model';

describe('TenantList - Delete action consistency', () => {
  let fixture: ComponentFixture<TenantList>;
  let overlayContainer: OverlayContainer;

  const mockTenant: Tenant = {
    id: 1,
    name: 'Jane Doe',
    tenant_type: 'Individual',
    phone_number: '01700000000',
    email: 'jane@example.com',
    nid_number: '1234567890',
    active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  beforeEach(async () => {
    const tenantServiceSpy = jasmine.createSpyObj<TenantService>('TenantService', [
      'getAllTenants',
    ]);
    tenantServiceSpy.getAllTenants.and.returnValue(
      of({
        tenants: [mockTenant],
        pagination: {
          current_page: 1,
          page_size: 100,
          total_items: 1,
          total_pages: 1,
          has_next: false,
          has_prev: false,
        },
      } as TenantListResponse)
    );

    const permissionServiceSpy = jasmine.createSpyObj<PermissionService>('PermissionService', [
      'isSuperAdmin',
      'isOrgAdmin',
      'isAdmin',
      'isPropertyManager',
      'isAccountant',
    ]);
    permissionServiceSpy.isSuperAdmin.and.returnValue(true);
    permissionServiceSpy.isOrgAdmin.and.returnValue(false);
    permissionServiceSpy.isAdmin.and.returnValue(false);
    permissionServiceSpy.isPropertyManager.and.returnValue(false);
    permissionServiceSpy.isAccountant.and.returnValue(false);

    const snackBarSpy = jasmine.createSpyObj<MatSnackBar>('MatSnackBar', ['open']);

    await TestBed.configureTestingModule({
      imports: [TenantList, TranslateModule.forRoot()],
      providers: [
        provideNoopAnimations(),
        { provide: TenantService, useValue: tenantServiceSpy },
        { provide: PermissionService, useValue: permissionServiceSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(TenantList);
    overlayContainer = TestBed.inject(OverlayContainer);
    fixture.detectChanges();
  });

  afterEach(() => {
    overlayContainer.ngOnDestroy();
  });

  function openRowMenu(): HTMLElement {
    const trigger = fixture.debugElement.query(By.directive(MatMenuTrigger));
    trigger.nativeElement.click();
    fixture.detectChanges();
    return overlayContainer.getContainerElement();
  }

  it('renders the delete action with the shared "delete" icon (matching Property/Lease)', () => {
    const overlayEl = openRowMenu();
    const deleteItem = overlayEl.querySelector('.delete-action');

    expect(deleteItem).withContext('delete menu item should exist').not.toBeNull();
    expect(deleteItem?.querySelector('mat-icon')?.textContent?.trim()).toBe('delete');
  });

  it('marks the delete action with the shared red "delete-action" class', () => {
    const overlayEl = openRowMenu();
    const deleteItem = overlayEl.querySelector('.delete-action');

    expect(deleteItem)
      .withContext('delete menu item should carry delete-action class')
      .not.toBeNull();
  });
});
