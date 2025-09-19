import { MetaDataResult, MetaData, ApiResult, RegionData } from "@/types";

/**
 * 查询meta信息
 */
export function getMeta(): Promise<MetaData> {
  return new Promise<MetaData>((resolve, reject) => {
    fetch("/meta").then(async (res) => {
      if (res.ok) {
        const json = await res.json();
        const data = json as MetaDataResult;
        console.log("load meta", data);

        if (data.code != 0) {
          reject(data.message);
        } else {
          resolve(data.data);
        }
      } else {
        reject(res.statusText);
      }
    });
  });
}

/**
 * 添加加速IP
 * @param ip 需要加速的IP
 * @param region 区域
 */
export function addIp(ip: string, region: string): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    fetch("/ip/add", {
      method: "post",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ ip: ip, region: region }),
    })
      .then(async (res) => {
        if (res.ok) {
          const result = await res.json();
          const { code, message } = result;
          if (code === 0) {
            resolve();
          } else {
            reject(message);
          }
        } else {
          reject(res.statusText);
        }
      })
      .catch((err) => {
        reject(err.message || err);
      });
  });
}

/**
 * 删除加速IP
 * @param ip 加速IP
 * @param region 区域
 */
export function deleteIp(ip: string, region: string): Promise<void> {
  return new Promise<void>((resolve, reject) => {
    fetch("/ip/delete", {
      method: "delete",
      body: JSON.stringify({
        region: region,
        ip: ip,
      }),
      headers: {
        "Content-Type": "application/json",
      },
    })
      .then(async (res) => {
        if (res.ok) {
          const result = await res.json();
          const { code, message } = result;
          if (code === 0) {
            resolve();
          } else {
            reject(message);
          }
        } else {
          reject(res.statusText);
        }
      })
      .catch((err) => {
        reject(err.message || err);
      });
  });
}

/**
 * 查询ip列表
 * @param region 区域
 */
export function listIp(region?: string): Promise<RegionData> {
  return new Promise<RegionData>((resolve, reject) => {
    fetch("/ip/list" + (region ? "?region=" + region : "")).then(
      async (res) => {
        if (res.ok) {
          const json = await res.json();
          const data = json as ApiResult<RegionData>;
          console.log("load ip list " + region, data);

          if (data.code != 0) {
            reject(data.message);
          } else {
            resolve(data.data);
          }
        } else {
          reject(res.statusText);
        }
      }
    );
  });
}
