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
        docs: {
          path: 'docs',
          routeBasePath: 'docs',
          sidebarPath: './sidebars.ts',
          lastVersion: process.env.STARPAC_DOCS_LAST_VERSION ?? 'current',
          versions: {
            current: {
              label: 'Next',
            },
          },
          editUrl: 'https://github.com/MagnusOpera/starpac/tree/main/website/',
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
      title: 'Starpac',
      logo: {
        alt: 'Starpac logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          to: '/docs/pgpac/',
          position: 'left',
          label: 'pgpac',
        },
        {
          to: '/docs/d1pac/',
          position: 'left',
          label: 'd1pac',
        },
        {
          type: 'docsVersionDropdown',
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
            {label: 'Installation', to: '/docs/pgpac/learn/installation'},
            {label: 'Quickstart', to: '/docs/pgpac/learn/quickstart'},
            {label: 'Reference', to: '/docs/pgpac/reference/project-file'},
          ],
        },
        {
          title: 'd1pac',
          items: [
            {label: 'Installation', to: '/docs/d1pac/learn/installation'},
            {label: 'Quickstart', to: '/docs/d1pac/learn/quickstart'},
            {label: 'Reference', to: '/docs/d1pac/reference/project-file'},
          ],
        },
        {
          title: 'Project',
          items: [
            {label: 'Repository', href: 'https://github.com/MagnusOpera/starpac'},
            {label: 'Changelog', href: 'https://github.com/MagnusOpera/starpac/blob/main/CHANGELOG.md'},
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
