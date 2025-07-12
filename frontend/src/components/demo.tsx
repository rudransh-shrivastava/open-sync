import { Button } from "@/components/ui/button";
import { useTheme } from "@/hooks/useTheme";

export default function Demo() {
  const { setTheme, theme } = useTheme();
  return (
    <div className="bg-background text-foreground flex h-svh w-full items-center justify-center gap-4 pb-4 text-4xl">
      Let's sync files!{" "}
      <Button
        onClick={() => {
          if (theme === "light") {
            setTheme("dark");
          } else {
            setTheme("light");
          }
        }}
      >
        <span className="dark:hidden">Switch To Dark</span>
        <span className="hidden dark:inline">Switch To Light</span>
      </Button>
    </div>
  );
}
