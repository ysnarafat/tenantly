import { Injectable, inject } from '@angular/core'; // Wait, inject is from @angular/core
import { Actions, createEffect, ofType } from '@ngrx/effects';
import { of } from 'rxjs';
import { map, mergeMap, catchError } from 'rxjs/operators';
import { BuildingService } from '../../../core/services/building.service';
import { BuildingActions } from './building.actions';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Building, BuildingListResponse } from '../../../core/models';
import { notifySuccess, notifyError } from '../../../shared/utils/notify.utils';

@Injectable()
export class BuildingEffects {
  private actions$ = inject(Actions);
  private buildingService = inject(BuildingService);
  private snackBar = inject(MatSnackBar);

  loadBuildings$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.loadBuildings),
      mergeMap(({ propertyId, active }) => {
        if (propertyId) {
          return this.buildingService.getBuildingsByProperty(propertyId).pipe(
            map((response: BuildingListResponse) =>
              BuildingActions.loadBuildingsSuccess({ buildings: response.buildings })
            ),
            catchError((error) => of(BuildingActions.loadBuildingsFailure({ error })))
          );
        } else {
          return this.buildingService.getBuildings({ active }).pipe(
            map((response: BuildingListResponse) =>
              BuildingActions.loadBuildingsSuccess({ buildings: response.buildings })
            ),
            catchError((error) => of(BuildingActions.loadBuildingsFailure({ error })))
          );
        }
      })
    )
  );

  loadBuilding$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.loadBuilding),
      mergeMap(({ id }) =>
        this.buildingService.getBuilding(id).pipe(
          map((building: Building) => BuildingActions.loadBuildingSuccess({ building })),
          catchError((error) => of(BuildingActions.loadBuildingFailure({ error })))
        )
      )
    )
  );

  createBuilding$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.createBuilding),
      mergeMap(({ request }) =>
        this.buildingService.createBuilding(request).pipe(
          map((building: Building) => {
            notifySuccess(this.snackBar, 'Building created successfully');
            return BuildingActions.createBuildingSuccess({ building });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Error creating building');
            return of(BuildingActions.createBuildingFailure({ error }));
          })
        )
      )
    )
  );

  updateBuilding$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.updateBuilding),
      mergeMap(({ id, request }) =>
        this.buildingService.updateBuilding(id, request).pipe(
          map((building: Building) => {
            notifySuccess(this.snackBar, 'Building updated successfully');
            return BuildingActions.updateBuildingSuccess({ building });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Error updating building');
            return of(BuildingActions.updateBuildingFailure({ error }));
          })
        )
      )
    )
  );

  deleteBuilding$ = createEffect(() =>
    this.actions$.pipe(
      ofType(BuildingActions.deleteBuilding),
      mergeMap(({ id }) =>
        this.buildingService.deleteBuilding(id).pipe(
          map(() => {
            notifySuccess(this.snackBar, 'Building deleted successfully');
            return BuildingActions.deleteBuildingSuccess({ id });
          }),
          catchError((error) => {
            notifyError(this.snackBar, 'Error deleting building');
            return of(BuildingActions.deleteBuildingFailure({ error }));
          })
        )
      )
    )
  );
}
