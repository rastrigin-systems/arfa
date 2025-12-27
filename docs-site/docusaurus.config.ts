import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Arfa Documentation',
  tagline: 'AI Agent Security Platform - Architecture & Design Decisions',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://docs.arfa.dev',
  baseUrl: '/',

  organizationName: 'rastrigin-systems',
  projectName: 'arfa',

  onBrokenLinks: 'warn',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/rastrigin-systems/arfa/tree/main/docs-site/',
          routeBasePath: '/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Arfa',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'architectureSidebar',
          position: 'left',
          label: 'Architecture',
        },
        {
          href: 'https://github.com/rastrigin-systems/arfa',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'Architecture Overview',
              to: '/',
            },
            {
              label: 'API Decisions',
              to: '/api/overview',
            },
            {
              label: 'CLI Decisions',
              to: '/cli/overview',
            },
          ],
        },
        {
          title: 'Resources',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/rastrigin-systems/arfa',
            },
            {
              label: 'OpenAPI Spec',
              href: 'https://github.com/rastrigin-systems/arfa/blob/main/platform/api-spec/spec.yaml',
            },
          ],
        },
      ],
      copyright: `Copyright ${new Date().getFullYear()} Arfa. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['go', 'bash', 'yaml', 'json'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
