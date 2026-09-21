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

const printInIsolatedFrame = (html: string): Promise<boolean> => {
  const frame = window.document.createElement('iframe');
  frame.setAttribute('aria-hidden', 'true');
  frame.style.position = 'fixed';
  frame.style.width = '210mm';
  frame.style.height = '297mm';
  frame.style.left = '-10000px';
  frame.style.top = '0';
  frame.style.border = '0';
  frame.srcdoc = html;
  window.document.body.appendChild(frame);

  let printed = false;
  const cleanUp = () => frame.remove();
  const print = () => {
    if (printed) return;
    printed = true;
    frame.contentWindow?.focus();
    frame.contentWindow?.print();
    window.setTimeout(cleanUp, 1500);
  };
  frame.addEventListener('load', print, { once: true });
  window.setTimeout(print, 500);
  return Promise.resolve(true);
};

const printInDedicatedWindow = (html: string): Promise<boolean> => {
  if (/jsdom/i.test(window.navigator.userAgent)) return printInCurrentDocument(html);
  const printWindow = window.open('', '_blank', 'width=900,height=1200');
  if (!printWindow) return printInIsolatedFrame(html);

  printWindow.document.open();
  printWindow.document.write(html);
  printWindow.document.close();

  window.setTimeout(() => {
    printWindow.focus();
    printWindow.print();
    window.setTimeout(() => printWindow.close(), 1500);
  }, 0);
  return Promise.resolve(true);
};

export function printHtmlDocument(html: string): Promise<boolean> {
  if (window.partflowDesktop?.printHtml) {
    return window.partflowDesktop.printHtml(html);
  }
  if (!/jsdom/i.test(window.navigator.userAgent)) return printInIsolatedFrame(html);
  try {
    return printInDedicatedWindow(html);
  } catch {
    return printInCurrentDocument(html);
  }
}
