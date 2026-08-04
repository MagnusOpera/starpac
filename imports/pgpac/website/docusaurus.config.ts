import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const baseUrl = '/pgpac/';

const config: Config = {
  title: 'pgpac',
  tagline: 'Package, diff, and apply PostgreSQL schema state with a Go-native CLI.',
  favicon: 'img/logo.svg',

  future: {
    v4: true,
  },

  url: 'https://magnusopera.github.io',
  baseUrl,

  organizationName: 'MagnusOpera',
  projectName: 'pgpac',

  onBrokenLinks: 'throw',

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
          routeBasePath: 'manual',
          lastVersion: process.env.PGPAC_DOCS_LAST_VERSION ?? 'current',
          versions: {
            current: {
              label: 'Next',
            },
          },
          editUrl: 'https://github.com/MagnusOpera/pgpac/tree/main/website/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/logo.svg',
    colorMode: {
      defaultMode: 'light',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'pgpac',
      logo: {
        alt: 'pgpac Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Docs',
        },
        {
          type: 'docsVersionDropdown',
          position: 'left',
          dropdownActiveClassDisabled: true,
        },
        {
          href: 'https://github.com/MagnusOpera/pgpac',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Installation',
              to: '/manual/learn/installation',
            },
            {
              label: 'Quickstart',
              to: '/manual/learn/quickstart',
            },
          ],
        },
        {
          title: 'Reference',
          items: [
            {
              label: 'Project File',
              to: '/manual/reference/project-file',
            },
            {
              label: 'Safety Model',
              to: '/manual/reference/safety-model',
            },
          ],
        },
        {
          title: 'Project',
          items: [
            {
              label: 'Repository',
              href: 'https://github.com/MagnusOpera/pgpac',
            },
            {
              label: 'Changelog',
              href: 'https://github.com/MagnusOpera/pgpac/blob/main/CHANGELOG.md',
            },
          ],
        },
      ],
      copyright: `© ${new Date().getFullYear()} <a class="footer-brand-link" href="https://magnusopera.io" target="_blank" rel="noopener noreferrer"><img src="${baseUrl}img/magnus-opera-logo.svg" alt="" />Magnus Opera SAS</a>`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['sql', 'bash', 'go', 'json'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
