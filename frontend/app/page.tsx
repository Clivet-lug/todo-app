"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { isLoggedIn } from "./lib/auth";

export default function RootPage() {
  const router = useRouter();

  useEffect(() => {
    if (isLoggedIn()) {
      router.replace("/dashboard");
    } else {
      router.replace("/login");
    }
  }, [router]);

  // Blank while redirecting
  return null;
}