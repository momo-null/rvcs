# Compliance & Responsible Use Assessment — RVCS

RVCS is a self-hosted remote device management system. This document assesses
the dual-use / privacy implications of its Android client (which includes remote
control capabilities) and outlines safeguards and deployment guidance.

## Capabilities of concern
1. Camera access & live video streaming (via LiveKit) — the device camera can be
   remotely activated and its feed viewed from the web console.
2. Microphone / two-way audio ("walkie-talkie") — remote audio capture.
3. Background / screen-off operation — a foreground service with WakeLock +
   WifiLock + watchdog keeps the app alive without ongoing user interaction.
4. Remote input injection — an Android `AccessibilityService`
   (see `android/app/src/main/res/xml/remote_accessibility_service.xml` and
   `docs/android-remote-control-plan.md`) can perform system-level actions
   (Back / Home / Recents, and other accessibility actions) on the device.

## Why this is dual-use
The combination of background operation, camera/microphone streaming, and remote
input injection is functionally similar to "stalkerware" / remote access trojans
(RATs). Such tooling is legal only when used on devices the operator owns or has
explicit, informed consent to monitor. The same code used on a device without
the owner's consent is illegal in many jurisdictions and violates platform
policies.

## Platform policy
- Google Play explicitly prohibits apps whose primary purpose is surveillance,
  or that access camera/microphone/location/SMS/contacts without prominent
  disclosure and user consent (Play Developer Policy on "Stalkerware" and "User
  Data"). Distributing this client through Google Play would very likely be
  rejected. Self-hosting / sideloading (ADB or enterprise MDM) is the intended
  distribution path.

## Legal / regulatory
- Many regions regulate covert surveillance, unauthorized interception of
  communications, and computer misuse. Ensure you have a lawful basis (ownership
  or documented, informed consent) before deploying on any device.
- If used in an employment, caregiving, or similar context, provide clear notice
  to the device users.

## Recommended safeguards (implement before broad deployment)
1. Authentication & authorization: every remote action requires the
   JWT-authenticated web session; add per-action confirmation and audit logging.
2. Consent & notice: ship a first-run consent screen on the device explaining
   camera/microphone/accessibility usage; the AccessibilityService already
   requires explicit, manual enablement by the user.
3. Audit trail: log every remote command (who / when / what) server-side.
4. No covert mode: the foreground-service notification must stay visible; do not
   suppress it.
5. Limit scope: keep remote input injection disabled by default; gate it behind
   a clear, opt-in setting.
6. Data minimization: do not persist video/audio longer than necessary.

## Intended use
RVCS is provided for legitimate scenarios such as monitoring devices you own
(kiosks, family-care devices with consent, authorized fleet / MDM). It must not
be used to surveil or control devices without the owner's informed consent.

## Disclaimer / Limitation of liability
RVCS is provided for educational and technical-sharing purposes, "as is", without
warranty of any kind, express or implied.
The author does not endorse, and is not responsible for, any use of this software
by third parties — including any deployment on a device without its owner's
informed consent. Users are solely responsible for ensuring their use of RVCS
complies with applicable laws and platform policies, and for any consequences
thereof. Nothing in this document constitutes legal advice.
