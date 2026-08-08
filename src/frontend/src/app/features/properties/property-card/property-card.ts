import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { Property, Building, Unit, UnitType, UnitWithDetails } from '../../../core/models';

export interface DisplayedProperty extends Property {
  expanded: boolean;
  buildings?: BuildingWithUnits[];
}

export interface BuildingWithUnits extends Building {
  expanded: boolean;
  units?: UnitWithDetails[];
}

@Component({
  selector: 'app-property-card',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatIconModule,
    MatButtonModule,
    MatChipsModule,
    MatTooltipModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './property-card.html',
  styleUrls: ['./property-card.scss'],
})
export class PropertyCardComponent {
  @Input() property!: DisplayedProperty;

  @Output() toggleProperty = new EventEmitter<DisplayedProperty>();
  @Output() editProperty = new EventEmitter<DisplayedProperty>();
  @Output() deleteProperty = new EventEmitter<DisplayedProperty>();
  @Output() addBuilding = new EventEmitter<DisplayedProperty>();
  @Output() editBuilding = new EventEmitter<{
    building: BuildingWithUnits;
    property: DisplayedProperty;
  }>();
  @Output() toggleBuilding = new EventEmitter<BuildingWithUnits>();
  @Output() addUnit = new EventEmitter<{
    building: BuildingWithUnits;
    property: DisplayedProperty;
  }>();
  @Output() editUnit = new EventEmitter<{
    unit: UnitWithDetails;
    building: BuildingWithUnits;
    property: DisplayedProperty;
  }>();
  @Output() deleteUnit = new EventEmitter<{ unit: UnitWithDetails; building: BuildingWithUnits }>();
  @Output() addPayment = new EventEmitter<{
    unit: UnitWithDetails;
    building: BuildingWithUnits;
    property: DisplayedProperty;
  }>();
  @Output() viewBuildingDetails = new EventEmitter<BuildingWithUnits>();
  @Output() viewUnitDetails = new EventEmitter<{
    unit: UnitWithDetails;
    building: BuildingWithUnits;
  }>();

  onToggleProperty() {
    this.toggleProperty.emit(this.property);
  }

  onEditProperty() {
    this.editProperty.emit(this.property);
  }

  onDeleteProperty() {
    this.deleteProperty.emit(this.property);
  }

  onAddBuilding() {
    this.addBuilding.emit(this.property);
  }

  onEditBuilding(building: BuildingWithUnits, event: Event) {
    event.stopPropagation();
    this.editBuilding.emit({ building, property: this.property });
  }

  onToggleBuilding(building: BuildingWithUnits) {
    this.toggleBuilding.emit(building);
  }

  onAddUnit(building: BuildingWithUnits) {
    this.addUnit.emit({ building, property: this.property });
  }

  onEditUnit(unit: UnitWithDetails, building: BuildingWithUnits) {
    this.editUnit.emit({ unit, building, property: this.property });
  }

  onDeleteUnit(unit: UnitWithDetails, building: BuildingWithUnits) {
    this.deleteUnit.emit({ unit, building });
  }

  onAddPayment(unit: UnitWithDetails, building: BuildingWithUnits) {
    this.addPayment.emit({ unit, building, property: this.property });
  }

  onViewBuildingDetails(building: BuildingWithUnits, event: Event) {
    event.stopPropagation();
    this.viewBuildingDetails.emit(building);
  }

  onViewUnitDetails(unit: UnitWithDetails, building: BuildingWithUnits) {
    this.viewUnitDetails.emit({ unit, building });
  }

  getPropertyTypeColor(type: string): string {
    switch (type) {
      case 'Residential':
        return 'primary';
      case 'Commercial':
        return 'accent';
      case 'Mixed':
        return 'warn';
      default:
        return '';
    }
  }

  getUnitTypeIcon(type: UnitType): string {
    switch (type) {
      case 'Shop':
        return 'store';
      case 'Apartment':
        return 'home';
      case 'Office':
        return 'business';
      case 'Parking':
        return 'local_parking';
      case 'Storage':
        return 'inventory_2';
      default:
        return 'meeting_room';
    }
  }
}
