(function(){
    let outputData = {};
    let selectedTab = document.querySelectorAll('#visKTTabset .h-tab-btns li.selected span.title');
    const findValueObject = (e) => {
        do {
            e = e.parentNode
            if (e.tagName.toLowerCase() === 'body') {
                return null;
            }
        } while (!e.classList.contains('colLabel'))
        e = e.parentNode;
        if (e.children.length === 0) {
            return null;
        }
        let valueNode = null;
        for(let node of e.children) {
            if (!node.classList.contains('colValue')) {
                continue;
            }
            valueNode = node
        }
        if (valueNode === null || valueNode.outerHTML.includes('<input')) {
            return null
        }
        return valueNode.innerText.trim();
    };
    const findTitle = (e) => {
        let container = e.parentNode.parentNode
        for(let hTag of container.children) {
            if (hTag.tagName.toLowerCase() !== 'h3' || !hTag.hasAttribute('title')) {
                continue;
            }
            return hTag.innerText.trim();
        }
        return null;
    };
    if (selectedTab.length === 0) {
        return {
            error: true,
            message: "Failed to find selected tab"
        }
    }
    selectedTab = selectedTab[0]
    let tab = 'unknown';
    if (selectedTab.innerHTML.includes('Køretøj')) {
        tab = 'vehicle';
    } else if (selectedTab.innerHTML.includes('Tekniske oplysninger')) {
        tab = 'technical_details';
    } else if (selectedTab.innerHTML.includes('Syn')) {
        tab = 'inspection';
    } else if (selectedTab.innerHTML.includes('Forsikring')) {
        tab = 'insurance';
    } else if (selectedTab.innerHTML.includes('tilladelser')) {
        tab = 'permissions';
    }
    if (tab === 'inspection') {
        outputData["never_inspected"] = document.body.innerHTML.includes('Køretøjet har aldrig været synet.');
        outputData["called_for_inspection"] = !(document.body.innerHTML.includes('Køretøjet er ikke indkaldt til syn.'))
    }
    if (["vehicle", "technical_details", "inspection"].includes(tab)) {
        let elementList = document.querySelectorAll('[id^="ptr-dmr:portlet"]');
        for(let element of elementList) {
            let outputKey = element.id
                .replaceAll(':', '_')
                .replaceAll('-', '_')
                .replaceAll('.', '_')
            if (outputKey.includes('HstrskVsnng')) {
                continue;
            }
            let outputValue = findValueObject(element)
            if (outputValue !== null) {
                outputData[outputKey] = outputValue
            }
        }
    }
    let identityElements = document.querySelectorAll('.bluebox .keyvalue')
    if (identityElements.length > 0) {
        let title = null;
        let key = null;
        let value = null;
        for(let node of identityElements) {
            key = null;
            value = null;
            title = findTitle(node);
            if (title === null || node.children.length === 0) {
                continue;
            }
            for(let span of node.children) {
                if (span.tagName.toLowerCase() === 'span' && span.classList.contains('key') && key === null) {
                    key = span.innerText.trim();
                }
                if (span.tagName.toLowerCase() === 'span' && span.classList.contains('value') && value === null) {
                    value = span.innerText.trim();
                }
            }
            if (key !== null && value !== null) {
                title = title.trim().replaceAll('­', '')
                key = key.replaceAll('­', '')
                    .replaceAll(':', '')
                    .replaceAll(',', '')
                    .replaceAll(/\s+/g, '_')
                if (typeof outputData[title] !== 'object') {
                    outputData[title] = {};
                }
                outputData[title][key] = value;
            }
        }
    }
    return outputData
})()