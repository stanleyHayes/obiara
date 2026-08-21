import { Box, Typography } from "@mui/material";
import Image from "next/image";
import type { ReactNode } from "react";

import brandMark from "../../../Obiara_Handover_Package/3_Brand/assets/logo/png/mark-color-ondark_transparent.png";

export function AuthShell({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <Box className="admin-auth-shell">
      <Box className="admin-auth-brand" component="aside">
        <div className="admin-auth-weave" aria-hidden="true" />
        <div className="admin-auth-brand-top">
          <Image
            alt=""
            className="admin-auth-mark"
            priority
            src={brandMark}
          />
          <div>
            <Typography className="admin-auth-wordmark">obiara</Typography>
            <Typography className="admin-auth-brand-label">
              Operations desk
            </Typography>
          </div>
        </div>

        <div className="admin-auth-statement">
          <Typography className="admin-auth-eyebrow">
            Stewardship, before access
          </Typography>
          <Typography component="p" className="admin-auth-quote">
            Every action here carries the weight of a community.
          </Typography>
          <Typography className="admin-auth-supporting-copy">
            A private workspace for the people entrusted with Obiara&apos;s
            safety, care and daily operations.
          </Typography>
        </div>

        <div className="admin-auth-trust-line">
          <span className="admin-auth-live-dot" aria-hidden="true" />
          <span>Protected operator environment</span>
        </div>
      </Box>

      <Box className="admin-auth-workspace" component="main">
        <div className="admin-auth-panel">{children}</div>
        <Typography className="admin-auth-footnote">
          Authorized Obiara operators only · Activity is recorded
        </Typography>
      </Box>
    </Box>
  );
}
