import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'database.dart';

/// Single shared [AppDatabase] instance for the whole app. Local-first: this
/// is the only source of truth right now — no network/auth/sync layer.
final appDatabaseProvider = Provider<AppDatabase>((ref) {
  final db = AppDatabase();
  ref.onDispose(db.close);
  return db;
});
