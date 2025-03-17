# Voc On Steroid - Front

## Setup

Make sure to install dependencies:

```bash
# bun
bun install
```

## Development Server

Start the development server on `http://localhost:3000`:

```bash
# bun
bun run dev
```

## Production

Build the application for production:

```bash
# bun
bun run build
```

Locally preview production build:

```bash
# bun
bun run preview
```

Check out the [deployment documentation](https://nuxt.com/docs/getting-started/deployment) for more information.

## Storybook

We use storybook for testing and showcasing component in isolation.

```bash
bun run storybook
```

## VSCode config

We use eslint for linting and formating. You should install the [ESlint extentions](https://marketplace.visualstudio.com/items?itemName=dbaeumer.vscode-eslint) and apply the following congig in your settings.json :

```json
  "eslint.validate": ["javascript", "javascriptreact", "typescript", "typescriptreact", "vue"],
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": "explicit",
    "source.fixAll.html": "explicit"
  },
  "eslint.format.enable": true,
  "editor.defaultFormatter": "dbaeumer.vscode-eslint",
  "prettier.enable": false,
  "eslint.options": {
    "overrideConfigFile": "eslint.config.mjs"
  },
```
