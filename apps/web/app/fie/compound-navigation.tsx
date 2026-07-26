import { fieRoutes, type FieRouteId } from "@obiara/fie-routing";
import Link from "next/link";

export type FieZone = Exclude<FieRouteId, "welcome" | "okyeame">;

const marks: Record<FieZone, string> = {
  home: "F",
  abonten: "A",
  adiwo: "D",
  "epono-ano": "Ɛ",
  "dan-mu": "M",
};

const navigation = fieRoutes
  .filter(
    (route): route is (typeof fieRoutes)[number] & { id: FieZone } =>
      route.id in marks,
  )
  .map((route) => ({
    label: route.label,
    gloss: route.gloss,
    zone: route.id,
    href: route.webPath,
    mark: marks[route.id],
  }));

interface CompoundNavigationProps {
  readonly current?: FieZone;
  readonly contextLabel?: string;
}

export function CompoundRail({
  current,
  contextLabel,
}: CompoundNavigationProps) {
  const currentLabel =
    contextLabel ?? navigation.find((item) => item.zone === current)?.label;

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
        <p>{currentLabel ?? "Fie"} is current</p>
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
