import path from "node:path";
import { fileURLToPath } from "node:url";
import ts from "typescript";

const packageDirectory = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const configPath = path.join(packageDirectory, "tsconfig.json");
const configFile = ts.readConfigFile(configPath, ts.sys.readFile);
if (configFile.error) {
  throw new Error(ts.flattenDiagnosticMessageText(configFile.error.messageText, "\n"));
}
const config = ts.parseJsonConfigFileContent(configFile.config, ts.sys, packageDirectory);
const program = ts.createProgram(config.fileNames, { ...config.options, noEmit: true });
const diagnostics = [...config.errors, ...ts.getPreEmitDiagnostics(program)];
if (diagnostics.length) {
  process.stderr.write(ts.formatDiagnosticsWithColorAndContext(diagnostics, {
    getCurrentDirectory: () => packageDirectory,
    getCanonicalFileName: fileName => fileName,
    getNewLine: () => "\n",
  }));
}

// TypeScript strict mode rejects implicit any but permits explicit any. Check
// syntax nodes as well so comments, strings, and generated dependencies are ignored.
const sourceFiles = new Set(config.fileNames.map(fileName => path.resolve(fileName)));
let explicitAnyCount = 0;
for (const source of program.getSourceFiles()) {
  if (!sourceFiles.has(path.resolve(source.fileName))) continue;
  function inspect(node) {
    if (node.kind === ts.SyntaxKind.AnyKeyword) {
      const position = source.getLineAndCharacterOfPosition(node.getStart(source));
      process.stderr.write(`${source.fileName}:${position.line + 1}:${position.character + 1}: Explicit any is not allowed; use a concrete type or narrow unknown.\n`);
      explicitAnyCount++;
    }
    ts.forEachChild(node, inspect);
  }
  inspect(source);
}
if (diagnostics.length || explicitAnyCount) process.exitCode = 1;
else process.stdout.write("SDK strict type checks and no-explicit-any check passed.\n");
