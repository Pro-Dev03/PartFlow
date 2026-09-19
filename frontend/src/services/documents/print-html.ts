const printInCurrentDocument = (html: string): Promise<boolean> => {
  const parsed = new DOMParser().parseFromString(html, 'text/html');
  const printRoot = window.document.createElement('div');
  const printId = `partflow-print-${Date.now()}`;
  printRoot.id = printId;
  printRoot.innerHTML = parsed.body.innerHTML;

  const printStyles = window.document.createElement('style');
  printStyles.dataset.partflowPrint = printId;
  printStyles.textContent = `${Array.from(parsed.head.querySelectorAll('style')).map((style) => style.textContent || '').join('\n')}\nbody > *:not(#${printId}){display:none!important}#${printId}{display:block!important}`;
  window.document.head.appendChild(printStyles);
  window.document.body.appendChild(printRoot);

  let cleanedUp = false;
  const cleanUp = () => {
    if (cleanedUp) return;
    cleanedUp = true;
    printRoot.remove();
    printStyles.remove();
    window.removeEventListener('afterprint', cleanUp);
  };

  window.addEventListener('afterprint', cleanUp, { once: true });
  window.setTimeout(cleanUp, 1500);
  try {
    window.print();
    return Promise.resolve(true);
  } catch (error) {
    cleanUp();
    return Promise.reject(error);
  }
};

export function printHtmlDocument(html: string): Promise<boolean> {
  if (window.partflowDesktop?.printHtml) {
    return window.partflowDesktop.printHtml(html);
  }
  return printInCurrentDocument(html);
}
