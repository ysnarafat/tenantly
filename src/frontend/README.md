# Tenantly Frontend

The frontend for the Tenantly property management platform, built with modern Angular.

## 🛠 Tech Stack

- **Framework**: Angular 20+ (standalone components)
- **Language**: TypeScript
- **Styling**: SCSS, Material Design
- **State Management**: NgRx (Store, Effects, Entity)
- **Build Tool**: Vite (Internal Dev Server) / esbuild (Build)

## 📁 Project Structure

```text
src/frontend/
├── src/
│   ├── app/              # Application core
│   │   ├── core/         # Services, guards, interceptors (Singleton logic)
│   │   ├── features/     # Feature modules (Business domains)
│   │   │   └── [feature]/
│   │   │       ├── [name].ts    # Standalone Component (Logic)
│   │   │       ├── [name].html  # Template
│   │   │       └── [name].scss  # Styles
│   │   └── shared/       # Shared UI components and pipes
│   ├── assets/           # Static assets
│   └── environments/     # Environment configurations
├── public/               # Public assets
└── angular.json          # Angular workspace configuration
```

## 📐 Component Conventions

Following modern Angular 20+ patterns:
- **Standalone Components**: No `NgModule` required.
- **File Naming**: Component files use the format `[name].ts` (without the `.component` suffix for the filename, e.g., `login.ts`).
- **External Templates**: Always use `templateUrl` and `styleUrls` to maintain clean separation.

## 🚀 Getting Started

### Prerequisites

- Node.js 25+
- npm 10+
- Angular CLI (`npm install -g @angular/cli`)

### Local Development

```bash
cd src/frontend
npm install
npm start
```
The app will be available at `http://localhost:4200`.

### Building for Production

```bash
npm run build
```
The build artifacts will be stored in the `dist/` directory.

## 🧪 Testing

```bash
# Run unit tests
npm test

# Run e2e tests
npm run e2e
```

## 🎨 Formatting & Quality

```bash
# Apply Prettier formatting
npm run format

# Run ESLint
npm run lint

# Check formatting (used in CI)
npm run format:check
```

## 🏗 CI/CD

This project uses GitHub Actions for Continuous Integration. Every push and pull request triggers the `Frontend CI` workflow, which ensures:
- All dependencies can be installed (`npm ci`).
- The code passes linting checks (`npm run lint`).
- The code follows formatting standards (`npm run format:check`).
