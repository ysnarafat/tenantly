import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/db_provider.dart';
import '../../core/db/tables.dart';

class PaymentWithDetails {
  PaymentWithDetails({
    required this.payment,
    required this.unitNumber,
    required this.tenantName,
  });

  final PaymentRow payment;
  final String unitNumber;
  final String tenantName;
}

/// Live query over Payments joined through their Lease to the Unit/Tenant,
/// same join-and-flatten pattern used for Leases.
final paymentsStreamProvider = StreamProvider<List<PaymentWithDetails>>((ref) {
  final db = ref.watch(appDatabaseProvider);
  final query =
      db.select(db.payments).join([
        innerJoin(db.leases, db.leases.id.equalsExp(db.payments.leaseId)),
        innerJoin(db.units, db.units.id.equalsExp(db.leases.unitId)),
        innerJoin(db.tenants, db.tenants.id.equalsExp(db.leases.tenantId)),
      ])..orderBy([
        OrderingTerm.desc(db.payments.year),
        OrderingTerm.desc(db.payments.month),
      ]);

  return query.watch().map(
    (rows) => rows
        .map(
          (row) => PaymentWithDetails(
            payment: row.readTable(db.payments),
            unitNumber: row.readTable(db.units).unitNumber,
            tenantName: row.readTable(db.tenants).name,
          ),
        )
        .toList(),
  );
});

final paymentByIdProvider = StreamProvider.family<PaymentRow?, int>((ref, id) {
  final db = ref.watch(appDatabaseProvider);
  return (db.select(
    db.payments,
  )..where((t) => t.id.equals(id))).watchSingleOrNull();
});

class PaymentsRepository {
  PaymentsRepository(this._db);

  final AppDatabase _db;

  Future<int> create({
    required int leaseId,
    required int month,
    required int year,
    required double amountDue,
    double amountPaid = 0,
    required PaymentStatus status,
    String? paymentMethod,
    DateTime? paymentDate,
    String? notes,
  }) {
    return _db
        .into(_db.payments)
        .insert(
          PaymentsCompanion.insert(
            leaseId: leaseId,
            month: month,
            year: year,
            amountDue: amountDue,
            amountPaid: Value(amountPaid),
            status: status,
            paymentMethod: Value(paymentMethod),
            paymentDate: Value(paymentDate),
            notes: Value(notes),
          ),
        );
  }

  Future<void> update(PaymentRow payment) {
    return _db
        .update(_db.payments)
        .replace(payment.copyWith(updatedAt: DateTime.now()));
  }

  Future<void> delete(int id) {
    return (_db.delete(_db.payments)..where((t) => t.id.equals(id))).go();
  }
}

final paymentsRepositoryProvider = Provider<PaymentsRepository>((ref) {
  return PaymentsRepository(ref.watch(appDatabaseProvider));
});
