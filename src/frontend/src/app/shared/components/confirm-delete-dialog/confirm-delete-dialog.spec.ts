import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { ConfirmDeleteDialogComponent, ConfirmDeleteDialogData } from './confirm-delete-dialog';
import { ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatDialogModule } from '@angular/material/dialog';
import { TranslateModule } from '@ngx-translate/core';
import { CommonModule } from '@angular/common';

describe('ConfirmDeleteDialogComponent', () => {
  let component: ConfirmDeleteDialogComponent;
  let fixture: ComponentFixture<ConfirmDeleteDialogComponent>;
  let mockDialogRef: jasmine.SpyObj<MatDialogRef<ConfirmDeleteDialogComponent>>;

  const dialogData: ConfirmDeleteDialogData = {
    entityLabel: 'property',
    entityName: 'Skyline Residences',
  };

  beforeEach(async () => {
    mockDialogRef = jasmine.createSpyObj<MatDialogRef<ConfirmDeleteDialogComponent>>(
      'MatDialogRef',
      ['close']
    );

    await TestBed.configureTestingModule({
      imports: [
        ConfirmDeleteDialogComponent,
        CommonModule,
        ReactiveFormsModule,
        MatFormFieldModule,
        MatInputModule,
        MatButtonModule,
        MatIconModule,
        MatDialogModule,
        TranslateModule.forRoot(),
      ],
      providers: [
        { provide: MatDialogRef, useValue: mockDialogRef },
        { provide: MAT_DIALOG_DATA, useValue: dialogData },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(ConfirmDeleteDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should keep confirm disabled when input is empty', () => {
    expect(component.isMatch).toBe(false);
  });

  it('should keep confirm disabled on a partial/mismatched name', () => {
    component.confirmationInput.setValue('Skyline');
    expect(component.isMatch).toBe(false);
  });

  it('should be case-sensitive', () => {
    component.confirmationInput.setValue('skyline residences');
    expect(component.isMatch).toBe(false);
  });

  it('should enable confirm on an exact match', () => {
    component.confirmationInput.setValue('Skyline Residences');
    expect(component.isMatch).toBe(true);
  });

  it('should not close the dialog when confirm is triggered without a match', () => {
    component.confirmationInput.setValue('wrong name');
    component.onConfirm();

    expect(mockDialogRef.close).not.toHaveBeenCalled();
  });

  it('should close with true when confirm is triggered on an exact match', () => {
    component.confirmationInput.setValue('Skyline Residences');
    component.onConfirm();

    expect(mockDialogRef.close).toHaveBeenCalledWith(true);
  });

  it('should close with false on cancel regardless of input', () => {
    component.confirmationInput.setValue('Skyline Residences');
    component.onCancel();

    expect(mockDialogRef.close).toHaveBeenCalledWith(false);
  });
});
