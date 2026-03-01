import { Injectable, inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { of } from 'rxjs';
import { map, mergeMap, catchError, tap } from 'rxjs/operators';
import { PropertyService } from '../../../core/services/property.service';
import { PropertyActions } from './property.actions';
import { MatSnackBar } from '@angular/material/snack-bar';

@Injectable()
export class PropertyEffects {
  private actions$ = inject(Actions);
  private propertyService = inject(PropertyService);
  private snackBar = inject(MatSnackBar);

  loadProperties$ = createEffect(() =>
    this.actions$.pipe(
      ofType(PropertyActions.loadProperties),
      mergeMap(({ active }) =>
        this.propertyService.getProperties({ active }).pipe(
          map((response) =>
            PropertyActions.loadPropertiesSuccess({ properties: response.properties })
          ),
          catchError((error) => of(PropertyActions.loadPropertiesFailure({ error: error.message })))
        )
      )
    )
  );

  loadProperty$ = createEffect(() =>
    this.actions$.pipe(
      ofType(PropertyActions.loadProperty),
      mergeMap(({ id }) =>
        this.propertyService.getProperty(id).pipe(
          map((property) => PropertyActions.loadPropertySuccess({ property })),
          catchError((error) => of(PropertyActions.loadPropertyFailure({ error: error.message })))
        )
      )
    )
  );

  createProperty$ = createEffect(() =>
    this.actions$.pipe(
      ofType(PropertyActions.createProperty),
      mergeMap(({ property }) =>
        this.propertyService.createProperty(property).pipe(
          map((newProperty) => {
            this.snackBar.open('Property created successfully', 'Close', { duration: 3000 });
            return PropertyActions.createPropertySuccess({ property: newProperty });
          }),
          catchError((error) => {
            this.snackBar.open('Failed to create property', 'Close', { duration: 3000 });
            return of(PropertyActions.createPropertyFailure({ error: error.message }));
          })
        )
      )
    )
  );

  updateProperty$ = createEffect(() =>
    this.actions$.pipe(
      ofType(PropertyActions.updateProperty),
      mergeMap(({ id, property }) =>
        this.propertyService.updateProperty(id, property).pipe(
          map((updatedProperty) => {
            this.snackBar.open('Property updated successfully', 'Close', { duration: 3000 });
            return PropertyActions.updatePropertySuccess({ property: updatedProperty });
          }),
          catchError((error) => {
            this.snackBar.open('Failed to update property', 'Close', { duration: 3000 });
            return of(PropertyActions.updatePropertyFailure({ error: error.message }));
          })
        )
      )
    )
  );

  deleteProperty$ = createEffect(() =>
    this.actions$.pipe(
      ofType(PropertyActions.deleteProperty),
      mergeMap(({ id }) =>
        this.propertyService.deleteProperty(id).pipe(
          map(() => {
            this.snackBar.open('Property deleted successfully', 'Close', { duration: 3000 });
            return PropertyActions.deletePropertySuccess({ id });
          }),
          catchError((error) => {
            this.snackBar.open('Failed to delete property', 'Close', { duration: 3000 });
            return of(PropertyActions.deletePropertyFailure({ error: error.message }));
          })
        )
      )
    )
  );
}
