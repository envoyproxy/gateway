Added a `clockSkew` field to `JWTProvider` in `SecurityPolicy`, mapping to Envoy's `clock_skew_seconds` and controlling the tolerance applied to time-based claims (`exp`, `nbf`).
