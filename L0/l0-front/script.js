const idInput = document.querySelector('.id-input')
const idSubmitButton = document.querySelector('.id-submit-btn')

const orderInfoDiv = document.querySelector('.order-info-div')
const orderInfoDivParagraph = orderInfoDiv.querySelector('p')
const orderInfoDivTextarea = orderInfoDiv.querySelector('textarea')


idSubmitButton.addEventListener('click', () => {
    const orderUID = idInput.value
    idInput.value = ''
    if (orderUID) {
        // sending get request for an order with specific order_uid
        var xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = function() {
            if (xmlHttp.readyState == XMLHttpRequest.DONE && xmlHttp.status == 200) {
                orderInfoDivTextarea.style.visibility = 'hidden'
                orderInfoDivParagraph.style.visibility = 'hidden'
                if (xmlHttp.responseText) {
                    var JSONtext = xmlHttp.responseText
                    var JSONobj = JSON.parse(JSONtext)
                    var orderInfoJSON = JSON.stringify(JSONobj, undefined, 4)
                    orderInfoDivTextarea.value = orderInfoJSON
                    orderInfoDivTextarea.style.visibility = 'visible'

                    orderInfoDivParagraph.innerHTML = "Данные о заказе с ID '" + orderUID  + "':"
                } else {
                    orderInfoDivParagraph.innerHTML = "Заказ с ID '" + orderUID + "' не найден..."
                }
                orderInfoDivParagraph.style.visibility = 'visible'
            }
        }
        xmlHttp.open("GET", "/orders/" + orderUID, true)
        xmlHttp.send(null)
    }
})

