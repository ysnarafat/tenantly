import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../tenants/tenants_providers.dart';
import '../units/units_providers.dart';
import 'leases_providers.dart';

class LeaseFormScreen extends ConsumerStatefulWidget {
  const LeaseFormScreen({super.key, this.leaseId});

  final int? leaseId;

  @override
  ConsumerState<LeaseFormScreen> createState() => _LeaseFormScreenState();
}

class _LeaseFormScreenState extends ConsumerState<LeaseFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _rentController = TextEditingController();
  int? _unitId;
  int? _tenantId;
  DateTime _startDate = DateTime.now();
  bool _active = true;
  bool _initialized = false;
  bool _saving = false;

  bool get _isEditMode => widget.leaseId != null;

  @override
  void dispose() {
    _rentController.dispose();
    super.dispose();
  }

  void _populateFromExisting(LeaseRow lease) {
    if (_initialized) return;
    _rentController.text = lease.monthlyRent.toStringAsFixed(0);
    _unitId = lease.unitId;
    _tenantId = lease.tenantId;
    _startDate = lease.startDate;
    _active = lease.active;
    _initialized = true;
  }

  Future<void> _pickStartDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _startDate,
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );
    if (picked != null) setState(() => _startDate = picked);
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    if (_unitId == null || _tenantId == null) return;
    setState(() => _saving = true);

    final rent = double.tryParse(_rentController.text.trim()) ?? 0;
    final repo = ref.read(leasesRepositoryProvider);
    if (_isEditMode) {
      final existing = await ref.read(
        leaseByIdProvider(widget.leaseId!).future,
      );
      if (existing != null) {
        await repo.update(
          existing.copyWith(
            unitId: _unitId!,
            tenantId: _tenantId!,
            monthlyRent: rent,
            startDate: _startDate,
            active: _active,
          ),
        );
      }
    } else {
      await repo.create(
        unitId: _unitId!,
        tenantId: _tenantId!,
        monthlyRent: rent,
        startDate: _startDate,
        active: _active,
      );
    }

    if (mounted) Navigator.of(context).pop();
  }

  Future<void> _delete() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete lease'),
        content: const Text('This cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(leasesRepositoryProvider).delete(widget.leaseId!);
      if (mounted) Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isEditMode) {
      final leaseAsync = ref.watch(leaseByIdProvider(widget.leaseId!));
      return leaseAsync.when(
        loading: () =>
            const Scaffold(body: Center(child: CircularProgressIndicator())),
        error: (error, stackTrace) =>
            Scaffold(body: Center(child: Text('Failed to load: $error'))),
        data: (lease) {
          if (lease == null) {
            return const Scaffold(body: Center(child: Text('Not found')));
          }
          _populateFromExisting(lease);
          return _buildForm();
        },
      );
    }
    return _buildForm();
  }

  Widget _buildForm() {
    final unitsAsync = ref.watch(unitsStreamProvider);
    final tenantsAsync = ref.watch(tenantsStreamProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditMode ? 'Edit Lease' : 'Add Lease'),
        actions: [
          if (_isEditMode)
            IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: 'Delete lease',
              onPressed: _delete,
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            unitsAsync.when(
              loading: () => const LinearProgressIndicator(),
              error: (error, stackTrace) => const Text('Failed to load units'),
              data: (units) {
                if (units.isEmpty) {
                  return const Text('Add a unit first before a lease.');
                }
                _unitId ??= units.first.unit.id;
                return DropdownButtonFormField<int>(
                  initialValue: _unitId,
                  decoration: const InputDecoration(labelText: 'Unit'),
                  items: units
                      .map(
                        (u) => DropdownMenuItem(
                          value: u.unit.id,
                          child: Text(
                            '${u.buildingName} · ${u.unit.unitNumber}',
                          ),
                        ),
                      )
                      .toList(),
                  onChanged: (value) => setState(() => _unitId = value),
                  validator: (value) => value == null ? 'Required' : null,
                );
              },
            ),
            const SizedBox(height: 12),
            tenantsAsync.when(
              loading: () => const LinearProgressIndicator(),
              error: (error, stackTrace) =>
                  const Text('Failed to load tenants'),
              data: (tenants) {
                if (tenants.isEmpty) {
                  return const Text('Add a tenant first before a lease.');
                }
                _tenantId ??= tenants.first.id;
                return DropdownButtonFormField<int>(
                  initialValue: _tenantId,
                  decoration: const InputDecoration(labelText: 'Tenant'),
                  items: tenants
                      .map(
                        (t) =>
                            DropdownMenuItem(value: t.id, child: Text(t.name)),
                      )
                      .toList(),
                  onChanged: (value) => setState(() => _tenantId = value),
                  validator: (value) => value == null ? 'Required' : null,
                );
              },
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _rentController,
              decoration: const InputDecoration(labelText: 'Monthly rent'),
              keyboardType: TextInputType.number,
              validator: (value) =>
                  (value == null || double.tryParse(value.trim()) == null)
                  ? 'Enter a valid amount'
                  : null,
            ),
            const SizedBox(height: 12),
            ListTile(
              contentPadding: EdgeInsets.zero,
              title: const Text('Start date'),
              subtitle: Text(
                '${_startDate.year}-${_startDate.month.toString().padLeft(2, '0')}-${_startDate.day.toString().padLeft(2, '0')}',
              ),
              trailing: const Icon(Icons.calendar_today_outlined),
              onTap: _pickStartDate,
            ),
            SwitchListTile(
              contentPadding: EdgeInsets.zero,
              title: const Text('Active'),
              value: _active,
              onChanged: (value) => setState(() => _active = value),
            ),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: _saving ? null : _save,
              child: Text(_saving ? 'Saving...' : 'Save'),
            ),
          ],
        ),
      ),
    );
  }
}
