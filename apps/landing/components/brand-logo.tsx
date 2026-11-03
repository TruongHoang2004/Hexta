import Image from "next/image";
import Link from "next/link";

interface BrandLogoProps {
  className?: string;
  showText?: boolean;
  href?: string;
}

export function BrandLogo({ className = "w-32", showText = false, href = "/" }: BrandLogoProps) {
  const content = (
    <div className="flex items-center gap-2.5 select-none">
      <span
        role="img"
        aria-label="Hexta Logo"
        className={`relative inline-block aspect-[3.3/1] shrink-0 overflow-hidden align-middle ${className}`}
      >
        <Image
          src="/brand/hexta-light.png"
          alt="Hexta Logo"
          width={1983}
          height={793}
          sizes="240px"
          priority
          className="absolute left-1/2 top-1/2 h-auto w-[125%] max-w-none -translate-x-1/2 -translate-y-1/2 dark:hidden"
        />
        <Image
          src="/brand/hexta-dark.png"
          alt="Hexta Logo"
          width={1983}
          height={793}
          sizes="240px"
          priority
          className="absolute left-1/2 top-1/2 hidden h-auto w-[125%] max-w-none -translate-x-1/2 -translate-y-1/2 dark:block"
        />
      </span>
      {showText && (
        <span className="font-bold text-lg tracking-tight text-foreground">
          Hexta
        </span>
      )}
    </div>
  );

  if (href) {
    return (
      <Link href={href} className="inline-flex items-center transition-opacity hover:opacity-90">
        {content}
      </Link>
    );
  }

  return content;
}
