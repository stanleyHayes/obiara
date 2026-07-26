import Link from "next/link";

export type FieZone = "fie" | "abonten" | "adiwo" | "epono-ano" | "dan-mu";

const navigation: readonly {
  label: string;
  gloss: string;
  zone: FieZone;
  href: string;
  mark: string;
}[] = [
  { label: "Fie", gloss: "home", zone: "fie", href: "/fie", mark: "F" },
  {
    label: "Abɔnten",
    gloss: "street",
    zone: "abonten",
    href: "/fie/abonten",
    mark: "A",
  },
  {
    label: "Adiwo",
    gloss: "courtyard",
    zone: "adiwo",
    href: "/fie/adiwo",
    mark: "D",
  },
  {
    label: "Ɛpono ano",
    gloss: "doorway",
    zone: "epono-ano",
    href: "/fie/epono-ano",
    mark: "Ɛ",
  },
  {
    label: "Dan mu",
    gloss: "inner room",
    zone: "dan-mu",
    href: "/fie/dan-mu",
    mark: "M",
  },
];

interface CompoundNavigationProps {
  readonly current: FieZone;
}

export function CompoundRail({ current }: CompoundNavigationProps) {
  const currentLabel = navigation.find((item) => item.zone === current)?.label;

  return (
    <aside className="fie-rail">
      <Link className="fie-wordmark" href="/">
        obiara
      </Link>
      <nav aria-label="Compound navigation">
        {navigation.map((item) => (
          <Link
            aria-current={item.zone === current ? "page" : undefined}
            href={item.href}
            key={item.zone}
          >
            <span aria-hidden="true">{item.mark}</span>
            <strong>{item.label}</strong>
            <small>{item.gloss}</small>
          </Link>
        ))}
      </nav>
      <div className="fie-rail-note">
        <span />
        <p>{currentLabel} is current</p>
        <small>Last safe sync 2 min ago</small>
      </div>
    </aside>
  );
}

export function CompoundBottomNavigation({ current }: CompoundNavigationProps) {
  return (
    <nav className="fie-bottom-nav" aria-label="Mobile compound navigation">
      {navigation.map((item) => (
        <Link
          aria-current={item.zone === current ? "page" : undefined}
          href={item.href}
          key={item.zone}
        >
          <span aria-hidden="true">{item.mark}</span>
          {item.label}
        </Link>
      ))}
    </nav>
  );
}
