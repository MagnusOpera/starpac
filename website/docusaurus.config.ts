import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const baseUrl = '/starpac/';

const config: Config = {
  title: 'Starpac',
  tagline: 'Desired-state database delivery for PostgreSQL and Cloudflare D1.',
  favicon: 'img/logo.svg',

  future: {
    v4: true,
  },

  url: 'https://magnusopera.github.io',
  baseUrl,
  organizationName: 'MagnusOpera',
  projectName: 'starpac',
  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: false,
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  plugins: [
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'pgpac',
        path: 'docs/pgpac',
        routeBasePath: 'pgpac',
        sidebarPath: './sidebars/pgpac.ts',
        lastVersion: process.env.PGPAC_DOCS_LAST_VERSION ?? 'current',
        versions: {
          current: {
            label: 'Next',
          },
        },
        editUrl: 'https://github.com/MagnusOpera/starpac/tree/main/website/',
      },
    ],
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'd1pac',
        path: 'docs/d1pac',
        routeBasePath: 'd1pac',
        sidebarPath: './sidebars/d1pac.ts',
        lastVersion: process.env.D1PAC_DOCS_LAST_VERSION ?? 'current',
        versions: {
          current: {
            label: 'Next',
          },
        },
        editUrl: 'https://github.com/MagnusOpera/starpac/tree/main/website/',
      },
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
      title: 'Starpac',
      logo: {
        alt: 'Starpac logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          docsPluginId: 'pgpac',
          sidebarId: 'docs',
          position: 'left',
          label: 'pgpac',
        },
        {
          type: 'docsVersionDropdown',
          docsPluginId: 'pgpac',
          position: 'left',
          dropdownActiveClassDisabled: true,
        },
        {
          type: 'docSidebar',
          docsPluginId: 'd1pac',
          sidebarId: 'docs',
          position: 'left',
          label: 'd1pac',
        },
        {
          type: 'docsVersionDropdown',
          docsPluginId: 'd1pac',
          position: 'left',
          dropdownActiveClassDisabled: true,
        },
        {
          href: 'https://github.com/MagnusOpera/starpac',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'pgpac',
          items: [
            {label: 'Installation', to: '/pgpac/learn/installation'},
            {label: 'Quickstart', to: '/pgpac/learn/quickstart'},
            {label: 'Reference', to: '/pgpac/reference/project-file'},
          ],
        },
        {
          title: 'd1pac',
          items: [
            {label: 'Installation', to: '/d1pac/learn/installation'},
            {label: 'Quickstart', to: '/d1pac/learn/quickstart'},
            {label: 'Reference', to: '/d1pac/reference/project-file'},
          ],
        },
        {
          title: 'Project',
          items: [
            {label: 'Repository', href: 'https://github.com/MagnusOpera/starpac'},
            {label: 'pgpac changelog', href: 'https://github.com/MagnusOpera/starpac/blob/main/products/pgpac/CHANGELOG.md'},
            {label: 'd1pac changelog', href: 'https://github.com/MagnusOpera/starpac/blob/main/products/d1pac/CHANGELOG.md'},
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
