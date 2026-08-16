import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import 'features/buildings/building_form_screen.dart';
import 'features/buildings/buildings_list_screen.dart';
import 'features/leases/lease_form_screen.dart';
import 'features/leases/leases_list_screen.dart';
import 'features/payments/payment_form_screen.dart';
import 'features/payments/payments_list_screen.dart';
import 'features/properties/properties_list_screen.dart';
import 'features/properties/property_form_screen.dart';
import 'features/reports/reports_screen.dart';
import 'features/tenants/tenant_form_screen.dart';
import 'features/tenants/tenants_list_screen.dart';
import 'features/units/unit_form_screen.dart';
import 'features/units/units_list_screen.dart';

final _router = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(path: '/', builder: (context, state) => const HomeShell()),
    GoRoute(
      path: '/properties/new',
      builder: (context, state) => const PropertyFormScreen(),
    ),
    GoRoute(
      path: '/properties/:id/edit',
      builder: (context, state) => PropertyFormScreen(
        propertyId: int.parse(state.pathParameters['id']!),
      ),
    ),
    GoRoute(
      path: '/buildings/new',
      builder: (context, state) => const BuildingFormScreen(),
    ),
    GoRoute(
      path: '/buildings/:id/edit',
      builder: (context, state) => BuildingFormScreen(
        buildingId: int.parse(state.pathParameters['id']!),
      ),
    ),
    GoRoute(
      path: '/units/new',
      builder: (context, state) => const UnitFormScreen(),
    ),
    GoRoute(
      path: '/units/:id/edit',
      builder: (context, state) =>
          UnitFormScreen(unitId: int.parse(state.pathParameters['id']!)),
    ),
    GoRoute(
      path: '/tenants/new',
      builder: (context, state) => const TenantFormScreen(),
    ),
    GoRoute(
      path: '/tenants/:id/edit',
      builder: (context, state) =>
          TenantFormScreen(tenantId: int.parse(state.pathParameters['id']!)),
    ),
    GoRoute(
      path: '/leases/new',
      builder: (context, state) => const LeaseFormScreen(),
    ),
    GoRoute(
      path: '/leases/:id/edit',
      builder: (context, state) =>
          LeaseFormScreen(leaseId: int.parse(state.pathParameters['id']!)),
    ),
    GoRoute(
      path: '/payments/new',
      builder: (context, state) => const PaymentFormScreen(),
    ),
    GoRoute(
      path: '/payments/:id/edit',
      builder: (context, state) =>
          PaymentFormScreen(paymentId: int.parse(state.pathParameters['id']!)),
    ),
  ],
);

class TenantlyMobileApp extends StatelessWidget {
  const TenantlyMobileApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Tenantly',
      theme: ThemeData(colorSchemeSeed: Colors.indigo, useMaterial3: true),
      routerConfig: _router,
    );
  }
}

class _Section {
  const _Section(this.label, this.icon, this.screen);

  final String label;
  final IconData icon;
  final Widget screen;
}

/// Side-drawer shell switching between the app's sections — a drawer rather
/// than a bottom nav bar since there are now seven of them, and it mirrors
/// the web app's sidenav more closely than a crowded bottom bar would.
class HomeShell extends StatefulWidget {
  const HomeShell({super.key});

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  int _index = 0;

  static const _sections = [
    _Section('Properties', Icons.apartment_outlined, PropertiesListScreen()),
    _Section('Buildings', Icons.domain_outlined, BuildingsListScreen()),
    _Section('Units', Icons.meeting_room_outlined, UnitsListScreen()),
    _Section('Tenants', Icons.people_outline, TenantsListScreen()),
    _Section('Leases', Icons.description_outlined, LeasesListScreen()),
    _Section('Payments', Icons.payments_outlined, PaymentsListScreen()),
    _Section('Reports', Icons.bar_chart_outlined, ReportsScreen()),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          NavigationDrawer(
            selectedIndex: _index,
            onDestinationSelected: (value) => setState(() => _index = value),
            children: [
              const Padding(
                padding: EdgeInsets.fromLTRB(16, 16, 16, 8),
                child: Text(
                  'Tenantly',
                  style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                ),
              ),
              for (final section in _sections)
                NavigationDrawerDestination(
                  icon: Icon(section.icon),
                  label: Text(section.label, overflow: TextOverflow.ellipsis),
                ),
            ],
          ),
          const VerticalDivider(width: 1),
          Expanded(
            child: IndexedStack(
              index: _index,
              children: [for (final section in _sections) section.screen],
            ),
          ),
        ],
      ),
    );
  }
}
