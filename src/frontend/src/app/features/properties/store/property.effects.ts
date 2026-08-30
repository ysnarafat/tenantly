import { Injectable, inject } from '@angular/core';
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { of } from 'rxjs';
import { map, mergeMap, catchError } from 'rxjs/operators';
import { PropertyService } from '../../../core/services/property.service';
import { PropertyActions } from './property.actions';
import { MatSnackBar } from '@angular/material/snack-bar';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

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
            notifySuccess(this.snackBar, 'Property created successfully');
            return PropertyActions.createPropertySuccess({ property: newProperty });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Failed to create property');
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
            notifySuccess(this.snackBar, 'Property updated successfully');
            return PropertyActions.updatePropertySuccess({ property: updatedProperty });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Failed to update property');
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
            notifySuccess(this.snackBar, 'Property deleted successfully');
            return PropertyActions.deletePropertySuccess({ id });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Failed to delete property');
            return of(PropertyActions.deletePropertyFailure({ error: error.message }));
          })
        )
      )
    )
  );
}
