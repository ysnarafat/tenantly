import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { TranslateModule } from '@ngx-translate/core';
import { SupportCenter } from './support-center';
import { SupportService } from '../../../core/services/support.service';
import { SupportCase, SupportCaseMessage } from '../../../core/models/support-case.model';
import { hasPermission, Permission } from '../../../core/models/role.model';

describe('SupportCenter', () => {
  let fixture: ComponentFixture<SupportCenter>;
  let component: SupportCenter;
  let supportService: jasmine.SpyObj<SupportService>;

  const supportCase: SupportCase = {
    id: 21,
    subject: 'Unable to add a property',
    description: 'The save button fails.',
    status: 'open',
    priority: 'normal',
    created_at: '2026-11-08T10:00:00Z',
    updated_at: '2026-11-08T10:00:00Z',
  };

  beforeEach(async () => {
    supportService = jasmine.createSpyObj<SupportService>('SupportService', [
      'listCases',
      'getCase',
      'createCase',
      'addMessage',
    ]);
    supportService.listCases.and.returnValue(of({ cases: [], total: 0 }));
    supportService.getCase.and.returnValue(of(supportCase));

    await TestBed.configureTestingModule({
      imports: [SupportCenter, TranslateModule.forRoot()],
      providers: [{ provide: SupportService, useValue: supportService }],
    }).compileComponents();

    fixture = TestBed.createComponent(SupportCenter);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('creates a case without sending organization, status, or priority fields', () => {
    supportService.createCase.and.returnValue(of(supportCase));
    supportService.listCases.and.returnValue(of({ cases: [supportCase], total: 1 }));

    component.newCase = {
      subject: '  Unable to add a property  ',
      description: '  The save button fails.  ',
    };
    component.createCase();

    expect(supportService.createCase).toHaveBeenCalledOnceWith({
      subject: 'Unable to add a property',
      description: 'The save button fails.',
    });
    expect(component.selectedCase?.id).toBe(supportCase.id);
  });

  it('sends a reply and adds the returned message to the selected conversation', () => {
    const message: SupportCaseMessage = {
      id: 22,
      case_id: supportCase.id,
      author: 'Tenant Admin',
      body: 'Here is more detail.',
      created_at: '2026-11-08T10:05:00Z',
    };
    supportService.addMessage.and.returnValue(of(message));
    component.selectedCase = supportCase;
    component.replyBody = '  Here is more detail.  ';

    component.addReply();

    expect(supportService.addMessage).toHaveBeenCalledOnceWith(
      supportCase.id,
      'Here is more detail.'
    );
    expect(component.selectedCase?.messages).toEqual([message]);
    expect(component.replyBody).toBe('');
  });

  it('grants Support Center access only to ORG_ADMIN', () => {
    expect(hasPermission('ORG_ADMIN', Permission.ACCESS_SUPPORT_CENTER)).toBeTrue();
    expect(hasPermission('Admin', Permission.ACCESS_SUPPORT_CENTER)).toBeFalse();
    expect(hasPermission('SUPER_ADMIN', Permission.ACCESS_SUPPORT_CENTER)).toBeFalse();
  });
});
