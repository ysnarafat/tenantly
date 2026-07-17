import { ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { provideNoopAnimations } from '@angular/platform-browser/animations';
import { OverlayContainer } from '@angular/cdk/overlay';
import { MatMenuTrigger } from '@angular/material/menu';
import { of } from 'rxjs';
import { TranslateModule } from '@ngx-translate/core';
import { MatSnackBar } from '@angular/material/snack-bar';
import { LeaseList } from './lease-list';
import {
  LeaseService,
  LeaseWithDetails,
  LeaseListResponse,
} from '../../../core/services/lease.service';
import { AuthService } from '../../../core/services/auth.service';

describe('LeaseList - Delete action consistency', () => {
  let fixture: ComponentFixture<LeaseList>;
  let overlayContainer: OverlayContainer;

  const mockLease: LeaseWithDetails = {
    id: 1,
    unit_id: 1,
    tenant_id: 1,
    lease_type: 'Residential',
    start_date: new Date().toISOString(),
    end_date: null as unknown as string,
    duration_months: 12,
    monthly_rent: 10000,
    security_deposit: 0,
    organization_id: 1,
    active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    building_id: 1,
    property_id: 1,
    property_name: 'Test Property',
    building_name: 'Main Building',
    building_code: 'MAIN',
    unit_number: '101',
    unit_type: 'Apartment',
    tenant_name: 'Jane Doe',
    is_expired: false,
    days_remaining: 300,
  };

  beforeEach(async () => {
    const leaseServiceSpy = jasmine.createSpyObj<LeaseService>('LeaseService', ['getAllLeases']);
    leaseServiceSpy.getAllLeases.and.returnValue(
      of({
        leases: [mockLease],
        pagination: {
          current_page: 1,
          page_size: 100,
          total_items: 1,
          total_pages: 1,
          has_next: false,
          has_prev: false,
        },
      } as LeaseListResponse)
    );

    const authServiceSpy = jasmine.createSpyObj<AuthService>('AuthService', [
      'isAdmin',
      'isPropertyManager',
    ]);
    authServiceSpy.isAdmin.and.returnValue(true);
    authServiceSpy.isPropertyManager.and.returnValue(false);

    const snackBarSpy = jasmine.createSpyObj<MatSnackBar>('MatSnackBar', ['open']);

    await TestBed.configureTestingModule({
      imports: [LeaseList, TranslateModule.forRoot()],
      providers: [
        provideNoopAnimations(),
        { provide: LeaseService, useValue: leaseServiceSpy },
        { provide: AuthService, useValue: authServiceSpy },
        { provide: MatSnackBar, useValue: snackBarSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(LeaseList);
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

  it('renders the delete action with the shared "delete" icon (matching Property/Tenant)', () => {
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
