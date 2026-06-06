import { UserRole } from "../enums/user-role.enum"

export {}


declare global {
  interface CustomJwtSessionClaims {
    metadata: {
      role?: UserRole
      onboardingCompleted?: boolean
    }
  }
}