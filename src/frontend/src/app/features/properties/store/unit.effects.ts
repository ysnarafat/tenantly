import { Injectable, inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { of } from 'rxjs';
import { map, mergeMap, catchError } from 'rxjs/operators';
import { UnitService } from '../../../core/services/unit.service';
import { UnitActions } from './unit.actions';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Unit, UnitListResponse } from '../../../core/models';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

@Injectable()
export class UnitEffects {
  private actions$ = inject(Actions);
  private unitService = inject(UnitService);
  private snackBar = inject(MatSnackBar);

  loadUnits$ = createEffect(() =>
    this.actions$.pipe(
      ofType(UnitActions.loadUnits),
      mergeMap(({ buildingId, active }) => {
        if (buildingId) {
          return this.unitService.getUnitsByBuilding(buildingId).pipe(
            map((response: UnitListResponse) =>
              UnitActions.loadUnitsSuccess({ units: response.units })
            ),
            catchError((error) => of(UnitActions.loadUnitsFailure({ error })))
          );
        } else {
          return this.unitService.getUnits({ active }).pipe(
            map((response: UnitListResponse) =>
              UnitActions.loadUnitsSuccess({ units: response.units })
            ),
            catchError((error) => of(UnitActions.loadUnitsFailure({ error })))
          );
        }
      })
    )
  );

  loadUnit$ = createEffect(() =>
    this.actions$.pipe(
      ofType(UnitActions.loadUnit),
      mergeMap(({ id }) =>
        this.unitService.getUnit(id).pipe(
          map((unit: Unit) => UnitActions.loadUnitSuccess({ unit })),
          catchError((error) => of(UnitActions.loadUnitFailure({ error })))
        )
      )
    )
  );

  createUnit$ = createEffect(() =>
    this.actions$.pipe(
      ofType(UnitActions.createUnit),
      mergeMap(({ request }) =>
        this.unitService.createUnit(request).pipe(
          map((unit: Unit) => {
            notifySuccess(this.snackBar, 'Unit created successfully');
            return UnitActions.createUnitSuccess({ unit });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Error creating unit');
            return of(UnitActions.createUnitFailure({ error }));
          })
        )
      )
    )
  );

  updateUnit$ = createEffect(() =>
    this.actions$.pipe(
      ofType(UnitActions.updateUnit),
      mergeMap(({ id, request }) =>
        this.unitService.updateUnit(id, request).pipe(
          map((unit: Unit) => {
            notifySuccess(this.snackBar, 'Unit updated successfully');
            return UnitActions.updateUnitSuccess({ unit });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Error updating unit');
            return of(UnitActions.updateUnitFailure({ error }));
          })
        )
      )
    )
  );

  deleteUnit$ = createEffect(() =>
    this.actions$.pipe(
      ofType(UnitActions.deleteUnit),
      mergeMap(({ id }) =>
        this.unitService.deleteUnit(id).pipe(
          map(() => {
            notifySuccess(this.snackBar, 'Unit deleted successfully');
            return UnitActions.deleteUnitSuccess({ id });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Error deleting unit');
            return of(UnitActions.deleteUnitFailure({ error }));
          })
        )
      )
    )
  );
}
