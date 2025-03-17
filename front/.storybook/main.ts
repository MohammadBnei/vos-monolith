import type { StorybookConfig } from '@storybook-vue/nuxt'

const config: StorybookConfig = {
  stories: [
    '../components/**/*.mdx',
    '../components/**/*.stories.@(js|jsx|ts|tsx|mdx)',
  ],
  addons: [
    '@storybook/addon-essentials',
    '@chromatic-com/storybook',
    '@storybook/experimental-addon-test',
    {
      name: '@storybook/addon-postcss',
      options: {
        postcssLoaderOptions: {
          // eslint-disable-next-line @typescript-eslint/no-require-imports
          implementation: require('postcss'),
        },
      },
    },
  ],
  framework: {
    name: '@storybook-vue/nuxt',
    options: {},
  },
}
export default config
