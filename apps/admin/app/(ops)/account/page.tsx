import { Suspense } from "react";

import { AccountSettings } from "./account-settings";

export default function AccountPage() {
  return (
    <Suspense>
      <AccountSettings />
    </Suspense>
  );
}
