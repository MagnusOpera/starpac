import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import styles from './index.module.css';

const workflow = [
  {
    number: '01',
    verb: 'Describe',
    title: 'Write the database you want',
    text: 'Keep the desired schema as ordinary SQL. Database changes become readable source-control changes.',
    commands: ['pgpac build', 'd1pac build'],
  },
  {
    number: '02',
    verb: 'Compare',
    title: 'See the exact path forward',
    text: 'Compare the immutable package with the live database and produce an ordered, inspectable plan.',
    commands: ['pgpac plan', 'd1pac plan'],
  },
  {
    number: '03',
    verb: 'Converge',
    title: 'Update reality to match',
    text: 'Apply the reviewed delta. Potentially destructive operations stay behind explicit safety gates.',
    commands: ['pgpac apply', 'd1pac apply'],
  },
];

const qualities = [
  ['SQL stays SQL', 'No proprietary schema language. Your database definition remains readable by people and its native engine.'],
  ['Drift is visible', 'The live target is part of every comparison, so a plan starts from reality—not an assumption.'],
  ['CI-friendly by design', 'Build once, inspect text or JSON plans, then promote the same package through environments.'],
];

function ArrowIcon() {
  return (
    <svg viewBox="0 0 20 20" aria-hidden="true">
      <path d="M4 10h11M11 6l4 4-4 4" />
    </svg>
  );
}

export default function Home() {
  return (
    <Layout
      title="Desired-state delivery for PostgreSQL and D1"
      description="Starpac is home to pgpac and d1pac: independent desired-state database delivery tools with one familiar workflow."
    >
      <main className={styles.page}>
        <section className={styles.hero}>
          <div className={`container ${styles.heroGrid}`}>
            <div className={styles.heroCopy}>
              <div className={styles.eyebrow}>
                <span className={styles.statusDot} />
                pgpac + d1pac
              </div>
              <Heading as="h1" className={styles.heroTitle}>
                Define the state.
                <span>Ship the difference.</span>
              </Heading>
              <p className={styles.heroLead}>
                Starpac is the home of two focused desired-state compilers. Use pgpac for PostgreSQL or d1pac
                for Cloudflare D1, with distinct packages and engine-native behavior under one Starpac version.
              </p>
              <div className={styles.actions}>
                <Link className={styles.primaryAction} to="/docs/pgpac/learn/quickstart">
                  Start with pgpac <ArrowIcon />
                </Link>
                <Link className={styles.primaryAction} to="/docs/d1pac/learn/quickstart">
                  Start with d1pac <ArrowIcon />
                </Link>
              </div>
              <p className={styles.dacpacNote}>
                <span>Coming from SQL Server?</span> Think DACPAC-style intent, built around each database's native SQL.
              </p>
            </div>

            <div className={styles.productVisual} aria-label="Starpac desired-state workflow example">
              <div className={styles.visualGlow} />
              <div className={styles.sourceCard}>
                <div className={styles.cardBar}>
                  <span>desired/</span>
                  <span className={styles.barMeta}>SQL source</span>
                </div>
                <div className={styles.fileTree}>
                  <div><span className={styles.treeLine} />Tables/accounts.sql</div>
                  <div><span className={styles.treeLine} />Views/active_accounts.sql</div>
                  <div><span className={styles.treeLine} />Indexes/accounts_status.sql</div>
                </div>
                <pre className={styles.sqlPreview}><code><span>CREATE TABLE</span> accounts (
  id integer <em>PRIMARY KEY</em>,
  status text
);</code></pre>
              </div>

              <div className={styles.compileRail}>
                <div className={styles.railLine} />
                <div className={styles.railBadge}>*pac plan</div>
                <div className={styles.railArrow}>↓</div>
              </div>

              <div className={styles.planCard}>
                <div className={styles.cardBar}>
                  <span>UPDATE PLAN</span>
                  <span className={styles.ready}><i /> READY</span>
                </div>
                <div className={styles.planRow}>
                  <span className={styles.add}>+</span>
                  <span>create table account_events</span>
                  <span className={styles.planOrder}>01</span>
                </div>
                <div className={styles.planRow}>
                  <span className={styles.change}>~</span>
                  <span>alter table accounts</span>
                  <span className={styles.planOrder}>02</span>
                </div>
                <div className={styles.planRow}>
                  <span className={styles.add}>+</span>
                  <span>create index accounts_status</span>
                  <span className={styles.planOrder}>03</span>
                </div>
                <div className={styles.planFooter}>
                  <span>desired</span><ArrowIcon /><span>live</span><strong>3 operations</strong>
                </div>
              </div>
            </div>
          </div>
          <div className={styles.heroRule} />
        </section>

        <section className={styles.processSection}>
          <div className="container">
            <div className={styles.sectionHeading}>
              <p className={styles.kicker}>The Starpac loop</p>
              <Heading as="h2">From desired state to database update.</Heading>
            </div>
            <div className={styles.workflow}>
              {workflow.map((item) => (
                <article className={styles.workflowStep} key={item.number}>
                  <div className={styles.stepTop}>
                    <span className={styles.stepNumber}>{item.number}</span>
                    <span className={styles.stepVerb}>{item.verb}</span>
                  </div>
                  <Heading as="h3">{item.title}</Heading>
                  <p>{item.text}</p>
                  <div className={styles.commandPair}>
                    {item.commands.map((command) => <code key={command}>{command}</code>)}
                  </div>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className={styles.ideaSection}>
          <div className={`container ${styles.ideaGrid}`}>
            <div className={styles.ideaCopy}>
              <p className={styles.kicker}>A familiar idea, engine-native execution</p>
              <Heading as="h2">Manage intent, not a pile of instructions.</Heading>
              <p>
                If you know DACPAC, the mental model will feel familiar: package a declarative database model,
                compare it to a target, then publish the delta.
              </p>
              <Link className={styles.textLink} to="/docs/pgpac/">
                Explore pgpac documentation <ArrowIcon />
              </Link>
            </div>
            <div className={styles.stateDiagram}>
              <div className={styles.stateBox}>
                <small>IN GIT</small><strong>Desired state</strong><span>versioned SQL</span>
              </div>
              <div className={styles.deltaBox}><span>compare</span><strong>Δ</strong><span>review</span></div>
              <div className={`${styles.stateBox} ${styles.liveState}`}>
                <small>DATABASE</small><strong>Live state</strong><span>updated safely</span>
              </div>
            </div>
          </div>
        </section>

        <section className={styles.qualitiesSection}>
          <div className="container">
            <div className={styles.qualitiesGrid}>
              {qualities.map(([title, text], index) => (
                <article className={styles.quality} key={title}>
                  <span>0{index + 1}</span><Heading as="h3">{title}</Heading><p>{text}</p>
                </article>
              ))}
            </div>
            <div className={styles.finalCta}>
              <div>
                <p className={styles.kicker}>Your schema already has a destination.</p>
                <Heading as="h2">Make it the source of truth.</Heading>
              </div>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
