import { Injectable, inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { of } from 'rxjs';
import { map, mergeMap, catchError } from 'rxjs/operators';
import { UnitService } from '../../../core/services/unit.service';
import { UnitActions } from './unit.actions';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Unit, UnitListResponse } from '../../../core/models';

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
                        map((response: UnitListResponse) => UnitActions.loadUnitsSuccess({ units: response.units })),
                        catchError(error => of(UnitActions.loadUnitsFailure({ error })))
                    );
                } else {
                    return this.unitService.getUnits({ active }).pipe(
                        map((response: UnitListResponse) => UnitActions.loadUnitsSuccess({ units: response.units })),
                        catchError(error => of(UnitActions.loadUnitsFailure({ error })))
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
                    catchError(error => of(UnitActions.loadUnitFailure({ error })))
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
                        this.snackBar.open('Unit created successfully', 'Close', { duration: 3000 });
                        return UnitActions.createUnitSuccess({ unit });
                    }),
                    catchError(error => {
                        this.snackBar.open('Error creating unit', 'Close', { duration: 3000 });
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
                        this.snackBar.open('Unit updated successfully', 'Close', { duration: 3000 });
                        return UnitActions.updateUnitSuccess({ unit });
                    }),
                    catchError(error => {
                        this.snackBar.open('Error updating unit', 'Close', { duration: 3000 });
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
                        this.snackBar.open('Unit deleted successfully', 'Close', { duration: 3000 });
                        return UnitActions.deleteUnitSuccess({ id });
                    }),
                    catchError(error => {
                        this.snackBar.open('Error deleting unit', 'Close', { duration: 3000 });
                        return of(UnitActions.deleteUnitFailure({ error }));
                    })
                )
            )
        )
    );
}
