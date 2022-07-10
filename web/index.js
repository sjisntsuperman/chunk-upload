// 10 MB
const chunkSize = 50 * 1024 * 1024 

// q1: 这里可能会序列号错乱，server 处理时需要注意
async function upload() {
    console.log(1111111111111111111)
    const file = document.getElementById("file").files[0]
    let start = 0, end = 0
    const fileSize = file.size
    const fileName = file.name
    let idx = 0
    const chunkTotal = Math.ceil(fileSize / chunkSize)
    while (start < fileSize) {
        end = start + chunkSize
        if (end > fileSize) {
            end = fileSize
        }
        var chunk = file.slice(start,end);//切割文件
        const formdata = new FormData()
        formdata.append("file", chunk, fileName)
        formdata.append("fileSize", fileSize)
        formdata.append("chunkIndex", idx++)
        formdata.append("chunkTotal", chunkTotal)
        await fetch('http://172.31.233.174:8001/merge', {
            method: "POST",
            mode: "no-cors",
            body: formdata
        }).then(res => {
            console.log(res)
        }).catch(err => {
            console.error(err)
        })
        start = end
    }
}