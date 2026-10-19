import Image from "next/image";

export function BrandLogo({ className = "w-32" }: { className?: string }) {
  return (
    <span
      role="img"
      aria-label="Hexta"
      className={`relative inline-block aspect-[3.3/1] shrink-0 overflow-hidden align-middle ${className}`}
    >
      {/* Frame the artwork without the transparent margins in the source files. */}
      <Image
        src="/brand/hexta-light.png"
        alt=""
        width={1983}
        height={793}
        sizes="240px"
        className="absolute left-1/2 top-1/2 h-auto w-[125%] max-w-none -translate-x-1/2 -translate-y-1/2 dark:hidden"
      />
      <Image
        src="/brand/hexta-dark.png"
        alt=""
        width={1983}
        height={793}
        sizes="240px"
        className="absolute left-1/2 top-1/2 hidden h-auto w-[125%] max-w-none -translate-x-1/2 -translate-y-1/2 dark:block"
      />
    </span>
  );
}
