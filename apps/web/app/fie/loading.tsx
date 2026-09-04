export default function FieLoading() {
  return (
    <main
      className="fie-page-skeleton"
      aria-label="Loading page"
      aria-busy="true"
    >
      <aside className="fie-skeleton-rail" aria-hidden="true">
        <i className="is-wordmark" />
        {Array.from({ length: 5 }, (_, index) => (
          <i key={index} />
        ))}
      </aside>
      <section>
        <div className="fie-skeleton-topbar" />
        <div className="fie-skeleton-hero">
          <div>
            <i />
            <i />
            <i />
          </div>
          <aside />
        </div>
        <div className="fie-skeleton-grid">
          {Array.from({ length: 3 }, (_, index) => (
            <i key={index} />
          ))}
        </div>
      </section>
      <span className="sr-only">Loading</span>
    </main>
  );
}
