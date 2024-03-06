const checkboxProcRe = /^check-([a-zA-Z\-]+)$/;
/**
 * returns the selected processes list as a string array from local storage
 *
 * @returns {String[]} the processes list
 */
const getSelectedProcs = () => {
  const selectedProcs =
    JSON.parse(localStorage.getItem("selected-procs")) ?? [];
  // nothing guarding this lol
  return selectedProcs;
};

/**
 * returns the list of process checkboxes as an object with key proc name
 *
 * @returns {Object.<string, HTMLInputElement>} a dictionary of key proc name and value checkbox
 */
const getProcCheckboxesByProc = () => {
  const checkboxes = document.getElementsByClassName("proc-check");
  let result = [];
  for (let i = 0; i < checkboxes.length; i++) {
    const checkbox = checkboxes.item(i);
    result[i] = [checkbox.id.match(checkboxProcRe)[1], checkbox];
  }

  const checkboxByProcName = Object.fromEntries(result);
  return checkboxByProcName;
};

/**
 * handles a checkbox being checked - saves the list of selected procs
 * in local storage
 *
 * @param {HTMLInputElement} self the checkbox (`this` in the html attr)
 */
const onCheckboxCheck = (self) => {
  const selectedProcs = getSelectedProcs();
  const procId = self.id.match(checkboxProcRe)[1];
  if (self.checked) {
    const newSelectedProcs = [...selectedProcs, procId];
    const distinctProcs = newSelectedProcs.filter(
      (proc, i) => newSelectedProcs.indexOf(proc) === i,
    );
    localStorage.setItem("selected-procs", JSON.stringify(distinctProcs));
    return;
  }

  const newSelectedProcs = selectedProcs.filter((proc) => proc != procId);
  localStorage.setItem("selected-procs", JSON.stringify(newSelectedProcs));
};

const registerCheckboxes = () => {
  const selectedProcs = getSelectedProcs();
  const procCheckboxes = getProcCheckboxesByProc();

  for (const i of selectedProcs) {
    procCheckboxes[i].checked = true;
  }

  for (const check of Object.values(procCheckboxes)) {
    check.onchange = () => onCheckboxCheck(check);
  }
};

const onload = () => {
  //document.addEventListener("htmx:afterSettle", (event) => {
  //  console.log({ event });
  //  registerCheckboxes();
  //});

  registerCheckboxes();
};

/**
 * clicks all the `className` elements under all selected procs (local storage)
 *
 * @param {String} className
 */
const clickAllButtonsUnderProc = (className) => {
  const selectedProcs = getSelectedProcs();
  const rows = selectedProcs
    .map((proc) => document.getElementById(`proc-${proc}`))
    // there may be nonexistent entries in the selected proc list
    .filter((elem) => elem);
  const startButtons = rows
    .map((row) => row.getElementsByClassName(className))
    .filter((collection) => collection.length === 1)
    .map((collection) => collection.item(0));

  for (const i of startButtons) {
    i.click();
  }
};

const startSelected = () => clickAllButtonsUnderProc("start-button");
const stopSelected = () => clickAllButtonsUnderProc("stop-button");
