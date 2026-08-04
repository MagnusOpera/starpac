import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import styles from './index.module.css';

const workflow = [
  {
    number: '01',
    verb: 'Describe',
    title: 'Write the database you want',
    text: 'Keep tables, views, functions, types, and permissions as standard SQL files. Comparing schema versions is as simple as reviewing a normal source-control diff.',
    command: 'pgpac build',
  },
  {
    number: '02',
    verb: 'Compare',
    title: 'See the exact path forward',
    text: 'pgpac compares the package with the live PostgreSQL database and produces an ordered, inspectable plan.',
    command: 'pgpac plan',
  },
  {
    number: '03',
    verb: 'Converge',
    title: 'Update reality to match',
    text: 'Apply the reviewed delta. Potentially destructive operations stay behind explicit safety gates.',
    command: 'pgpac apply',
  },
];

const qualities = [
  ['SQL stays SQL', 'No proprietary schema language. Your database definition remains readable by people and PostgreSQL.'],
  ['Drift is visible', 'The live target is always part of the comparison, so a plan starts from reality—not an assumption.'],
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
      title="Desired-state schema management for PostgreSQL"
      description="Define the PostgreSQL schema you want. Let pgpac package, compare, and safely update the live database to match."
    >
      <main className={styles.page}>
        <section className={styles.hero}>
          <div className={`container ${styles.heroGrid}`}>
            <div className={styles.heroCopy}>
              <div className={styles.eyebrow}>
                <span className={styles.statusDot} />
                Desired state for PostgreSQL
              </div>
              <Heading as="h1" className={styles.heroTitle}>
                Define the state.
                <span>Ship the difference.</span>
              </Heading>
              <p className={styles.heroLead}>
                pgpac is the desired-state compiler for PostgreSQL. Describe the schema you want in SQL,
                compare it with what is running, and apply a safe, reviewable update.
              </p>
              <div className={styles.actions}>
                <Link className={styles.primaryAction} to="/manual/learn/quickstart">
                  Build your first package <ArrowIcon />
                </Link>
                <Link className={styles.secondaryAction} to="/manual/">
                  Explore the docs
                </Link>
              </div>
              <p className={styles.dacpacNote}>
                <span>Coming from SQL Server?</span> Think DACPAC-style intent, built natively around PostgreSQL and SQL.
              </p>
            </div>

            <div className={styles.productVisual} aria-label="pgpac desired-state workflow example">
              <div className={styles.visualGlow} />
              <div className={styles.sourceCard}>
                <div className={styles.cardBar}>
                  <span>desired/</span>
                  <span className={styles.barMeta}>SQL source</span>
                </div>
                <div className={styles.fileTree}>
                  <div><span className={styles.treeLine} />Tables/accounts.sql</div>
                  <div><span className={styles.treeLine} />Views/active_accounts.sql</div>
                  <div><span className={styles.treeLine} />Functions/account_count.sql</div>
                </div>
                <pre className={styles.sqlPreview}><code><span>CREATE TABLE</span> accounts {'{'}
  id uuid <em>PRIMARY KEY</em>,
  status account_status
{'}'};</code></pre>
              </div>

              <div className={styles.compileRail}>
                <div className={styles.railLine} />
                <div className={styles.railBadge}>pgpac plan</div>
                <div className={styles.railArrow}>↓</div>
              </div>

              <div className={styles.planCard}>
                <div className={styles.cardBar}>
                  <span>UPDATE PLAN</span>
                  <span className={styles.ready}><i /> READY</span>
                </div>
                <div className={styles.planRow}>
                  <span className={styles.add}>+</span>
                  <span>create type account_status</span>
                  <span className={styles.planOrder}>01</span>
                </div>
                <div className={styles.planRow}>
                  <span className={styles.change}>~</span>
                  <span>alter table accounts</span>
                  <span className={styles.planOrder}>02</span>
                </div>
                <div className={styles.planRow}>
                  <span className={styles.add}>+</span>
                  <span>create view active_accounts</span>
                  <span className={styles.planOrder}>03</span>
                </div>
                <div className={styles.planFooter}>
                  <span>desired</span>
                  <ArrowIcon />
                  <span>live</span>
                  <strong>3 operations</strong>
                </div>
              </div>
            </div>
          </div>
          <div className={styles.heroRule} />
        </section>

        <section className={styles.processSection}>
          <div className="container">
            <div className={styles.sectionHeading}>
              <p className={styles.kicker}>The pgpac loop</p>
              <Heading as="h2">From desired state to database update.</Heading>
              <p>You own the destination. pgpac works out the journey.</p>
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
                  <code>{item.command}</code>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className={styles.ideaSection}>
          <div className={`container ${styles.ideaGrid}`}>
            <div className={styles.ideaCopy}>
              <p className={styles.kicker}>A familiar idea, a PostgreSQL workflow</p>
              <Heading as="h2">Manage intent, not a pile of instructions.</Heading>
              <p>
                If you know DACPAC, the mental model will feel familiar: package a declarative database model,
                compare it to a target, then publish the delta. pgpac brings that process to PostgreSQL without
                trying to make PostgreSQL behave like SQL Server.
              </p>
              <Link className={styles.textLink} to="/manual/learn/quickstart">
                See the complete workflow <ArrowIcon />
              </Link>
            </div>
            <div className={styles.stateDiagram}>
              <div className={styles.stateBox}>
                <small>IN GIT</small>
                <strong>Desired state</strong>
                <span>versioned SQL</span>
              </div>
              <div className={styles.deltaBox}>
                <span>compare</span>
                <strong>Δ</strong>
                <span>review</span>
              </div>
              <div className={`${styles.stateBox} ${styles.liveState}`}>
                <small>POSTGRESQL</small>
                <strong>Live state</strong>
                <span>updated safely</span>
              </div>
            </div>
          </div>
        </section>

        <section className={styles.qualitiesSection}>
          <div className="container">
            <div className={styles.qualitiesGrid}>
              {qualities.map(([title, text], index) => (
                <article className={styles.quality} key={title}>
                  <span>0{index + 1}</span>
                  <Heading as="h3">{title}</Heading>
                  <p>{text}</p>
                </article>
              ))}
            </div>
            <div className={styles.finalCta}>
              <div>
                <p className={styles.kicker}>Your schema already has a destination.</p>
                <Heading as="h2">Make it the source of truth.</Heading>
              </div>
              <Link className={styles.primaryAction} to="/manual/learn/installation">
                Install pgpac <ArrowIcon />
              </Link>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
